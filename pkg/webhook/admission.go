//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	goHttp "net/http"
	"reflect"
	"time"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func NewAdmissionHandler[T meta.Object](name, group, version, kind, resource string, handlers ...Handler[T]) Admission {
	return admissionImpl[T]{
		name:     name,
		group:    group,
		version:  version,
		kind:     kind,
		resource: resource,
		handlers: handlers,
	}
}

type AdmissionRequestType int

const (
	AdmissionRequestValidate AdmissionRequestType = iota
	AdmissionRequestMutate
)

type Admissions []Admission

func (a Admissions) Register() util.Mod[goHttp.ServeMux] {
	return func(in *goHttp.ServeMux) {
		for _, handler := range a {
			log := logger.Str("name", handler.Name())

			log.Info("Registering handler")

			if endpoint := fmt.Sprintf("/webhook/%s/%s/validate", gvsAsPath(handler.Resource()), handler.Name()); endpoint != "" {
				log.Str("endpoint", endpoint).Info("Registered Validate handler")
				in.HandleFunc(endpoint, handler.Validate)
			}

			if endpoint := fmt.Sprintf("/webhook/%s/%s/mutate", gvsAsPath(handler.Resource()), handler.Name()); endpoint != "" {
				log.Str("endpoint", endpoint).Info("Registered Mutate handler")
				in.HandleFunc(endpoint, handler.Mutate)
			}
		}
	}
}

func gvsAsPath(in meta.GroupVersionResource) string {
	if in.Group == "" {
		return fmt.Sprintf("core/%s/%s", in.Version, in.Resource)
	}
	return fmt.Sprintf("%s/%s/%s", in.Group, in.Version, in.Resource)
}

type Admission interface {
	Name() string

	Kind() meta.GroupVersionKind
	Resource() meta.GroupVersionResource

	Validate(goHttp.ResponseWriter, *goHttp.Request)
	Mutate(goHttp.ResponseWriter, *goHttp.Request)
}

type admissionImpl[T meta.Object] struct {
	name, group, version, kind, resource string

	handlers []Handler[T]
}

func (a admissionImpl[T]) Name() string {
	return a.name
}

func (a admissionImpl[T]) Kind() meta.GroupVersionKind {
	return meta.GroupVersionKind{
		Group:   a.group,
		Version: a.version,
		Kind:    a.kind,
	}
}

func (a admissionImpl[T]) Resource() meta.GroupVersionResource {
	return meta.GroupVersionResource{
		Group:    a.group,
		Version:  a.version,
		Resource: a.resource,
	}
}

func (a admissionImpl[T]) Validate(writer goHttp.ResponseWriter, request *goHttp.Request) {
	a.request(AdmissionRequestValidate, writer, request)
}

func (a admissionImpl[T]) Mutate(writer goHttp.ResponseWriter, request *goHttp.Request) {
	a.request(AdmissionRequestMutate, writer, request)
}

func (a admissionImpl[T]) request(t AdmissionRequestType, writer goHttp.ResponseWriter, request *goHttp.Request) {
	log := logger.Wrap(logging.HTTPRequestWrap(request))

	log.Info("Request Received")

	timeout := time.Second

	if request.URL.Query().Has("timeout") {
		if v, err := time.ParseDuration(request.URL.Query().Get("timeout")); err == nil {
			if v > 500*time.Millisecond {
				timeout = v - 200*time.Millisecond
			} else {
				timeout = v
			}
		}
	}

	ctx, c := context.WithTimeout(shutdown.Context(), timeout)
	defer c()

	code, data := a.requestWriterJSON(ctx, log, t, request)
	writer.WriteHeader(code)
	if len(data) > 0 {
		if _, err := util.WriteAll(writer, data); err != nil {
			log.Err(err).Warn("Unable to send response")
		}
	}
}

func (a admissionImpl[T]) requestWriterJSON(ctx context.Context, log logging.Logger, t AdmissionRequestType, request *goHttp.Request) (int, []byte) {
	code, obj, err := a.requestWriter(ctx, log, t, request)

	if err != nil {
		if herr, ok := http.IsError(err); ok {
			return herr.Code, herr.JSON()
		}

		log.Err(err).Warn("Unexpected Error")
		return goHttp.StatusInternalServerError, nil
	}

	if reflect.ValueOf(obj).IsZero() {
		return code, nil
	}

	data, err := json.Marshal(obj)
	if err != nil {
		log.Err(err).Warn("Unable to marshal response")
		return goHttp.StatusInternalServerError, nil
	}

	return code, data
}

func (a admissionImpl[T]) requestWriter(ctx context.Context, log logging.Logger, t AdmissionRequestType, request *goHttp.Request) (int, any, error) {
	switch t {
	case AdmissionRequestValidate, AdmissionRequestMutate:
	default:
		log.Warn("Invalid AdmissionRequestType")
		return 0, nil, http.NewError(goHttp.StatusBadRequest, "Invalid AdmissionRequestType")
	}

	if request.Method != goHttp.MethodPost {
		return 0, nil, http.NewError(goHttp.StatusMethodNotAllowed, "Method '%s' not allowed, expected '%s'", request.Method, goHttp.MethodPost)
	}

	var req admission.AdmissionReview

	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		return 0, nil, http.WrapError(goHttp.StatusBadRequest, err)
	}

	resp := a.admissionHandle(ctx, log, t, req.Request)

	return 200, admission.AdmissionReview{
		TypeMeta: req.TypeMeta,
		Response: resp,
	}, nil
}

func (a admissionImpl[T]) admissionHandle(ctx context.Context, log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest) *admission.AdmissionResponse {
	if request == nil {
		return &admission.AdmissionResponse{
			Allowed: false,
		}
	}

	if request.Kind != a.Kind() {
		return &admission.AdmissionResponse{
			UID:     request.UID,
			Allowed: false,
			Result: &meta.Status{
				Message: fmt.Sprintf("Invalid Kind. Got '%s', expected '%s'", request.Kind.String(), a.Kind().String()),
			},
		}
	}

	resp, err := a.admissionHandleE(ctx, log, t, request)
	if err != nil {
		return &admission.AdmissionResponse{
			UID:     request.UID,
			Allowed: false,
			Result: &meta.Status{
				Message: fmt.Sprintf("Unexpected error: %s", err.Error()),
			},
		}
	}

	if resp == nil {
		return &admission.AdmissionResponse{
			UID:     request.UID,
			Allowed: false,
			Result: &meta.Status{
				Message: "Missing Response element",
			},
		}
	}

	resp.UID = request.UID

	return resp
}

func (a admissionImpl[T]) admissionHandleE(ctx context.Context, log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest) (*admission.AdmissionResponse, error) {
	old, err := a.evaluateObject(request.OldObject.Raw)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to parse old object")
	}

	new, err := a.evaluateObject(request.Object.Raw)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to parse new object")
	}

	switch t {
	case AdmissionRequestValidate:
		for _, handler := range a.handlers {
			if handler.CanHandle(ctx, log, t, request, old, new) {
				if v, ok := handler.(ValidationHandler[T]); ok {
					result, err := v.Validate(ctx, log, t, request, old, new)
					if err != nil {
						return nil, err
					}

					return result.AsResponse()
				}

				return ValidationResponse{
					Allowed: false,
					Message: "Request not handled",
				}.AsResponse()
			}
		}

		return &admission.AdmissionResponse{Allowed: true}, nil
	case AdmissionRequestMutate:
		for _, handler := range a.handlers {
			if handler.CanHandle(ctx, log, t, request, old, new) {
				if v, ok := handler.(MutationHandler[T]); ok {
					result, err := v.Mutate(ctx, log, t, request, old, new)
					if err != nil {
						return nil, err
					}

					return result.AsResponse()
				}

				return ValidationResponse{
					Allowed: false,
					Message: "Request not handled",
				}.AsResponse()
			}
		}

		return &admission.AdmissionResponse{Allowed: true}, nil
	default:
		return &admission.AdmissionResponse{
			Allowed: false,
		}, nil
	}
}

func (a admissionImpl[T]) evaluateObject(data []byte) (T, error) {
	if len(data) == 0 {
		return util.Default[T](), nil
	}

	obj, err := util.DeepType[T]()
	if err != nil {
		return util.Default[T](), err
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return util.Default[T](), err
	}

	return obj, nil
}
