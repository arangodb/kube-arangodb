//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"context"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
)

type AuthResponse struct {
	Username string

	Token *string

	CustomResponse *pbEnvoyAuthV3.CheckResponse

	Headers []*pbEnvoyCoreV3.HeaderValueOption
}

type AuthRequestFunc func(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *AuthResponse) (*AuthResponse, error)

func MergeAuthRequest(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, requests ...AuthRequestFunc) (*AuthResponse, error) {
	var resp *AuthResponse
	for _, r := range requests {
		if v, err := r(ctx, request, resp); err != nil {
			return nil, err
		} else {
			resp = v
		}
	}

	return resp, nil
}
