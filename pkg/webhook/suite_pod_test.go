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
	"testing"

	"github.com/stretchr/testify/require"
	admission "k8s.io/api/admission/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func newPodAdmissionRequest(t *testing.T, name, namespace string, op admission.Operation, old, new *core.Pod) *admission.AdmissionRequest {
	req := &admission.AdmissionRequest{
		UID: "",
		Kind: meta.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Pod",
		},
		Resource: meta.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		},
		Name:      name,
		Namespace: namespace,
		Operation: op,
	}

	if old != nil {
		data, err := json.Marshal(old)
		require.NoError(t, err)
		req.OldObject.Raw = data
	}

	if new != nil {
		data, err := json.Marshal(new)
		require.NoError(t, err)
		req.Object.Raw = data
	}

	return req
}

func newPodAdmission(name string, handlers ...Handler[*core.Pod]) Admission {
	return NewAdmissionHandler[*core.Pod](name, "", "v1", "Pod", "pods", handlers...)
}

func newPodHandler(can CanHandleFunc[*core.Pod],
	mutate MutateFunc[*core.Pod],
	validate ValidateFunc[*core.Pod]) Handler[*core.Pod] {
	return podHandler{
		can:      can,
		mutate:   mutate,
		validate: validate,
	}
}

type podHandler struct {
	can      CanHandleFunc[*core.Pod]
	mutate   MutateFunc[*core.Pod]
	validate ValidateFunc[*core.Pod]
}

func (p podHandler) Validate(ctx context.Context, log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new *core.Pod) (ValidationResponse, error) {
	if p.validate == nil {
		return ValidationResponse{}, nil
	}

	return p.validate(ctx, log, t, request, old, new)
}

func (p podHandler) Mutate(ctx context.Context, log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new *core.Pod) (MutationResponse, error) {
	if p.mutate == nil {
		return MutationResponse{}, nil
	}

	return p.mutate(ctx, log, t, request, old, new)
}

func (p podHandler) CanHandle(ctx context.Context, log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new *core.Pod) bool {
	if p.can == nil {
		return false
	}

	return p.can(ctx, log, t, request, old, new)
}
