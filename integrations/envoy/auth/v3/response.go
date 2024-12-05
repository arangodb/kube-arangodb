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

package v3

import (
	"encoding/json"
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	status "google.golang.org/genproto/googleapis/rpc/status"
)

type DeniedMessage struct {
	Message string `json:"message,omitempty"`
}

type DeniedResponse struct {
	Code    int32
	Headers map[string]string
	Message *DeniedMessage
}

func (d DeniedResponse) Error() string {
	return fmt.Sprintf("Request denied with code: %d", d.Code)
}

func (d DeniedResponse) GetCheckResponse() (*pbEnvoyAuthV3.CheckResponse, error) {
	var resp pbEnvoyAuthV3.DeniedHttpResponse

	for k, v := range d.Headers {
		resp.Headers = append(resp.Headers, &corev3.HeaderValueOption{
			Header: &corev3.HeaderValue{
				Key:   k,
				Value: v,
			},
			AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
		})
	}

	if data := d.Message; data != nil {
		z, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		resp.Body = string(z)
		resp.Headers = append(resp.Headers, &corev3.HeaderValueOption{
			Header: &corev3.HeaderValue{
				Key:   "content-type",
				Value: "application/json",
			},
			AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
		})
	}

	resp.Status = &typev3.HttpStatus{
		Code: typev3.StatusCode(d.Code),
	}

	return &pbEnvoyAuthV3.CheckResponse{
		HttpResponse: &pbEnvoyAuthV3.CheckResponse_DeniedResponse{DeniedResponse: &resp},
		Status: &status.Status{
			Code: d.Code,
		}}, nil
}
