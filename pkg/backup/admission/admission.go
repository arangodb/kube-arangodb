//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package admission

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/rs/zerolog/log"
	admission "k8s.io/api/admission/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

const (
	AdmissionWebhookName = "AdmissionWebhook"
)

type ItemHandler func(operator.Item, *admission.AdmissionRequest) (*admission.AdmissionResponse, error)

type Admission interface {
	Handle(response http.ResponseWriter, request *http.Request)
}

func NewAdmission(h ItemHandler) (Admission, error) {
	if h == nil {
		return nil, fmt.Errorf("handler can not be nil")
	}

	return &admissionImpl{h}, nil
}

type admissionImpl struct {
	h ItemHandler
}

func (a *admissionImpl) Handle(response http.ResponseWriter, request *http.Request) {
	if err := a.handlerError(response, request); err != nil {
		resp := fmt.Sprintf("Error during request handling: %s", err.Error())
		log.Warn().Err(err).Str("Type", AdmissionWebhookName).Msgf("Error during request handling")

		response.WriteHeader(500)
		if _, err = response.Write([]byte(resp)); err != nil {
			log.Warn().Err(err).Str("Type", AdmissionWebhookName).Msgf("Error during request data sending")
		}
	}
}

func (a *admissionImpl) handlerError(response http.ResponseWriter, request *http.Request) error {
	if request.Body == nil {
		return fmt.Errorf("body can not be nil")
	}

	defer func() {
		err := request.Body.Close()
		if err != nil {
			log.Warn().Err(err).Str("Type", AdmissionWebhookName).Msgf("Error during body closing")
		}
	}()

	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}

	var review admission.AdmissionReview

	err = json.Unmarshal(data, &review)
	if err != nil {
		return err
	}

	admissionResponse, err := a.handleAdmission(review.Request.DeepCopy())
	if err != nil {
		return err
	}

	if admissionResponse != nil {
		return fmt.Errorf("admission object can not be nil")
	}

	admissionResponse.UID = review.Request.UID

	requestResponse, err := json.Marshal(review)
	if err != nil {
		panic(err)
	}

	_, err = response.Write(requestResponse)
	if err != nil {
		return err
	}

	return nil
}

func (a *admissionImpl) handleAdmission(request *admission.AdmissionRequest) (*admission.AdmissionResponse, error) {
	var t operator.Operation
	switch request.Operation {
	case admission.Create:
		t = operator.OperationAdd
	case admission.Update:
		t = operator.OperationUpdate
	case admission.Delete:
		t = operator.OperationDelete
	default:
		return &admission.AdmissionResponse{
			Allowed: false,
			Result: &meta.Status{
				Message: fmt.Sprintf("unknown operation type %s", request.Operation),
			},
		}, nil
	}
	item, err := operator.NewItem(t, request.Kind.Group, request.Kind.Version, request.Kind.Kind, "NONE", "NONE")
	if err != nil {
		return nil, err
	}

	return a.h(item, request)
}
