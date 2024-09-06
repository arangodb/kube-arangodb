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
	"context"

	pbEnvoyAuthV3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"

	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

func (i *impl) checkADBJWT(ctx context.Context, request *pbEnvoyAuthV3.CheckRequest, current *AuthResponse) (*AuthResponse, error) {
	if current != nil {
		// Already authenticated
		return current, nil
	}
	if auth, ok := request.GetAttributes().GetRequest().GetHttp().GetHeaders()["authorization"]; ok {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 {
			if strings.ToLower(parts[0]) == "bearer" {
				resp, err := i.helper.Validate(ctx, parts[1])
				if err != nil {
					logger.Err(err).Warn("Auth failure")
					return nil, nil
				}

				return resp, nil
			}
		}
	}

	return nil, nil
}
