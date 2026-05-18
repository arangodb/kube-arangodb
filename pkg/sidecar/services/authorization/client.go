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
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (a *implementation) Ready(ctx context.Context) error {
	return a.Health(ctx).Require()
}

func (a *implementation) Revision() uint64 {
	return uint64(a.policies.Index() + a.roles.Index())
}

func (a *implementation) Evaluate(ctx context.Context, req *pbAuthorizationV1.AuthorizationV1PermissionRequest) (*pbAuthorizationV1.AuthorizationV1PermissionResponse, error) {
	// Superuser: when no user and no roles are specified the caller is the
	// operator itself (authenticated via the internal JWT) — grant full access.
	if req.User == nil && len(req.GetGroups()) == 0 {
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
			Message: "Superuser access",
			Effect:  sidecarSvcAuthzTypes.Effect_Allow,
		}, nil
	}

	groups, err := a.getUserGroups(req.GetUser(), req.GetGroups()...)
	if err != nil {
		return nil, err
	}

	resp, err := groups.Evaluate(req)
	if err != nil {
		return nil, err
	}

	l := logger.Str("user", req.GetUser()).Str("action", req.GetAction()).Str("resource", req.GetResource())

	if resp.GetEffect() == sidecarSvcAuthzTypes.Effect_Deny {
		l.Str("reason", resp.GetMessage()).Info("Permission denied")
	} else {
		l.Debug("Permission granted")
	}

	return resp, nil
}

func (a *implementation) getUserGroups(user string, groupNames ...string) (sidecarSvcAuthzClient.ScopedPolicies, error) {
	allPolicies := a.policies.Copy()
	allGroups := a.roles.Copy()

	result := make(sidecarSvcAuthzClient.ScopedPolicies, len(groupNames))

	// Collect groups from explicit request
	for name, g := range allGroups {
		if !util.ContainsList(groupNames, name) {
			continue
		}

		if sp, err := a.resolveGroup(g, allPolicies); err != nil {
			return nil, err
		} else if sp != nil {
			result[name] = *sp
		}
	}

	// Collect groups from user bindings
	if user != "" {
		allBindings := a.userRoleBindings.Copy()
		prefix := user + ":"

		for key, binding := range allBindings {
			if len(key) <= len(prefix) || key[:len(prefix)] != prefix {
				continue
			}

			groupName := binding.GetRole()
			if _, exists := result[groupName]; exists {
				continue
			}

			if g, ok := allGroups[groupName]; ok {
				if sp, err := a.resolveGroup(g, allPolicies); err != nil {
					return nil, err
				} else if sp != nil {
					result[groupName] = *sp
				}
			}
		}
	}

	return result, nil
}

func (a *implementation) resolveGroup(g *sidecarSvcAuthzTypes.Role, allPolicies map[string]*sidecarSvcAuthzTypes.Policy) (*sidecarSvcAuthzClient.ScopedPolicy, error) {
	scope := g.GetScope()
	if scope == nil {
		return nil, nil
	}

	var sp sidecarSvcAuthzClient.ScopedPolicy

	if p, err := sidecarSvcAuthzClient.NewPolicy(scope); err != nil {
		return nil, err
	} else {
		sp.Scope = &p
	}

	for _, policyName := range g.Policies {
		if p, ok := allPolicies[policyName]; ok {
			if pol, err := sidecarSvcAuthzClient.NewPolicy(p); err != nil {
				return nil, err
			} else {
				sp.Policies = append(sp.Policies, &pol)
			}
		}
	}

	return &sp, nil
}
