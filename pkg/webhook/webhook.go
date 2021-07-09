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
	"encoding/json"
	"net/http"

	"github.com/arangodb/kube-arangodb/pkg/util/log"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	admission "k8s.io/api/admission/v1beta1"
)

func NewWebhookHandler(log zerolog.Logger, handler WebhookHandler) (http.Handler, error) {
	if handler == nil {
		return nil, errors.Errorf("Handler cannot be nil")
	}

	return webhook{
		log:     log,
		handler: handler,
	}, nil
}

type WebhookHandler interface {
	Handle(log log.Factory, request *admission.AdmissionRequest) admission.AdmissionResponse
}

var _ http.Handler = &webhook{}

type webhook struct {
	log zerolog.Logger

	handler WebhookHandler
}

func (w webhook) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	code, data := w.serveHTTP(response, request)

	if code != http.StatusOK {
		response.WriteHeader(code)
	}

	if data != nil {
		if _, writeErr := response.Write(data); writeErr != nil {
			w.log.Error().Err(writeErr).Msg("Webhook request failed - unable to write body")
		}
	}
}

func (w webhook) serveHTTP(response http.ResponseWriter, request *http.Request) (int, []byte) {
	if request.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, []byte("Only POST is allowed")
	}

	var review admission.AdmissionReview

	if err := json.NewDecoder(request.Body).Decode(&review); err != nil {
		w.log.Warn().Err(err).Msg("Unable to decode json body for admission review")
		return http.StatusInternalServerError, []byte(err.Error())
	}

	if review.Request == nil {
		return http.StatusInternalServerError, []byte("Invalid request send by API Server")
	}

	admissionResponse := w.review(review.Request.DeepCopy())
	admissionResponse.UID = review.Request.UID
	review.Response = admissionResponse.DeepCopy()

	// From this point we can return only 200
	if err := json.NewEncoder(response).Encode(&review); err != nil {
		w.log.Warn().Err(err).Msg("Unable to decode json body for admission review")
	}

	return http.StatusOK, nil
}

func (w webhook) review(request *admission.AdmissionRequest) admission.AdmissionResponse {
	log := wrapLogRequest(w.log, request)

	return w.handler.Handle(log, request)
}
