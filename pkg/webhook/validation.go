//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package webhook

import (
	"fmt"
	"net/http"

	"github.com/arangodb/kube-arangodb/pkg/util/log"
	"github.com/rs/zerolog"
	admission "k8s.io/api/admission/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewValidationWebhookHandler(log zerolog.Logger) http.Handler {
	hook, err := NewWebhookHandler(log, validation{})

	if err != nil {
		// This part will never be reached
		panic(err)
	}

	return hook
}

var _ WebhookHandler = &validation{}

type validation struct {
}

func (v validation) Handle(log log.Factory, request *admission.AdmissionRequest) admission.AdmissionResponse {
	log.Info().Msg("Validation request received")

	allowed, message := v.handle(log, request)

	response := admission.AdmissionResponse{
		Allowed: allowed,
	}

	if message != "" {
		response.Result = &meta.Status{
			Message: message,
		}
	}
	log.Info().Bool("allowed", allowed).Str("message", message).Msg("Validation request processed")

	return response
}

func (v validation) handle(log log.Factory, request *admission.AdmissionRequest) (bool, string) {
	matchingHandlers := handlers.Filter(request.Kind)

	if len(matchingHandlers) == 0 {
		return false, fmt.Sprintf("Unable to handle object %s", request.Kind.String())
	}

	switch request.Operation {
	case admission.Create:
		createHandlers := matchingHandlers.AsCreateHandler()

		if len(createHandlers) == 0 {
			return false, fmt.Sprintf("Unable to handle object CREATE operation %s", request.Kind.String())
		}

		for _, handler := range createHandlers {
			if allowed, message := handler.ValidateCreate(log, request.DeepCopy()); !allowed {
				return false, message
			}
		}

		return true, ""
	case admission.Update:
		updateHandlers := matchingHandlers.AsUpdateHandler()

		if len(updateHandlers) == 0 {
			return false, fmt.Sprintf("Unable to handle object UPDATE operation %s", request.Kind.String())
		}

		for _, handler := range updateHandlers {
			if allowed, message := handler.ValidateUpdate(log, request.DeepCopy()); !allowed {
				return false, message
			}
		}

		return true, ""
	case admission.Delete:
		deleteHandlers := matchingHandlers.AsDeleteHandler()

		if len(deleteHandlers) == 0 {
			return false, fmt.Sprintf("Unable to handle object DELETE operation %s", request.Kind.String())
		}

		for _, handler := range deleteHandlers {
			if allowed, message := handler.ValidateDelete(log, request.DeepCopy()); !allowed {
				return false, message
			}
		}

		return true, ""
	default:
		return false, fmt.Sprintf("Unknown operation %s", request.Operation)
	}
}
