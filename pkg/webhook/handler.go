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
	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/logging"
)

type CanHandleFunc[T meta.Object] func(log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new T) bool

type MutateFunc[T meta.Object] func(log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new T) (MutationResponse, error)

type ValidateFunc[T meta.Object] func(log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new T) (ValidationResponse, error)

type Handler[T meta.Object] interface {
	CanHandle(log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new T) bool
}

type MutationHandler[T meta.Object] interface {
	Handler[T]

	Mutate(log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new T) (MutationResponse, error)
}

type ValidationHandler[T meta.Object] interface {
	Handler[T]

	Validate(log logging.Logger, t AdmissionRequestType, request *admission.AdmissionRequest, old, new T) (ValidationResponse, error)
}
