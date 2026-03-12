//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package client

import (
	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
)

func EvaluatePolicies(req *pbAuthorizationV1.AuthorizationV1PermissionRequest, policies ...*Policy) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	context := req.GetContext().GetContext()

	var allowed bool

	for _, policy := range policies {
		if a, err := policy.Evaluate(req.GetAction(), req.GetResource(), context); err != nil {
			if sidecarSvcAuthzTypes.IsPermissionDenied(err) {
				return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
					Message: "Explicit deny",
					Effect:  pbAuthorizationV1.AuthorizationV1Effect_Deny,
				}, nil
			}
		} else if a {
			allowed = true
		}
	}

	if allowed {
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
			Message: "Access Granted",
			Effect:  pbAuthorizationV1.AuthorizationV1Effect_Allow,
		}, nil
	}
	return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
		Message: "Permission denied",
		Effect:  pbAuthorizationV1.AuthorizationV1Effect_Deny,
	}, nil
}
