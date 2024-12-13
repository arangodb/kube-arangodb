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
	"encoding/json"
	"fmt"

	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewValidationResponse(allowed bool, msg string, args ...any) ValidationResponse {
	return ValidationResponse{
		Allowed: allowed,
		Message: fmt.Sprintf(msg, args...),
	}
}

type ValidationResponse struct {
	Allowed  bool
	Message  string
	Warnings []string
}

func (v ValidationResponse) AsResponse() (*admission.AdmissionResponse, error) {
	if v.Allowed {
		return &admission.AdmissionResponse{
			Allowed:  true,
			Warnings: v.Warnings,
		}, nil
	}

	return &admission.AdmissionResponse{
		Allowed:  false,
		Warnings: v.Warnings,
		Result: &meta.Status{
			Message: v.Message,
		},
	}, nil
}

type MutationResponse struct {
	ValidationResponse

	Patch patch.Items
}

func (v MutationResponse) AsResponse() (*admission.AdmissionResponse, error) {
	resp, err := v.ValidationResponse.AsResponse()
	if err != nil {
		return nil, err
	}

	if len(v.Patch) == 0 {
		return resp, nil
	}

	q, err := json.Marshal(v.Patch)
	if err != nil {
		return nil, err
	}

	resp.Patch = q
	resp.PatchType = util.NewType(admission.PatchTypeJSONPatch)

	return resp, nil
}
