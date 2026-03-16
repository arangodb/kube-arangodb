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

package authorization

import (
	"context"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	sidecarSvcAuthzClient "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (a *implementation) Ready(ctx context.Context) error {
	return a.Health(ctx).Require()
}

func (a *implementation) Revision() uint64 {
	return uint64(a.policies.Index() + a.roles.Index())
}

func (a *implementation) Evaluate(ctx context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	policies, err := a.getUserPolicies(req.GetUser(), req.GetRoles()...)
	if err != nil {
		return nil, err
	}

	return sidecarSvcAuthzClient.EvaluatePolicies(req, policies...)
}

func (a *implementation) getUserPolicies(user string, roles ...string) ([]*sidecarSvcAuthzClient.Policy, error) {
	allPolicies := a.policies.Copy()
	allRoles := a.roles.Copy()

	var policyNames = make(map[string]*sidecarSvcAuthzClient.Policy, len(allPolicies))

	for k, v := range allRoles {
		if util.ContainsList(roles, k) || util.ContainsList(v.Users, user) {
			for _, name := range v.Policies {
				if p, ok := allPolicies[name]; ok {
					if _, ok := policyNames[name]; !ok {
						if v, err := sidecarSvcAuthzClient.NewPolicy(p); err != nil {
							return nil, err
						} else {
							policyNames[name] = &v
						}
					}
				}
			}
		}
	}

	return util.MapValues(policyNames), nil
}
