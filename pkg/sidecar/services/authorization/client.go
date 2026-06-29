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
	if req.User == nil && len(req.GetRoles()) == 0 {
		return &pbAuthorizationV1.AuthorizationV1PermissionResponse{
			Message: "Superuser access",
			Effect:  sidecarSvcAuthzTypes.Effect_Allow,
		}, nil
	}

	groups, err := a.getUserGroups(req.GetUser())
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

// getUserGroups resolves the scoped policies for a user from the user-role
// bindings only. The scope is taken from each binding; the (deprecated) role
// scope is intentionally ignored. Roles passed explicitly in the request are not
// considered - only roles bound to the user via an ArangoPermissionRoleUserBinding
// (or token) grant access.
func (a *implementation) getUserGroups(user string) (sidecarSvcAuthzClient.ScopedPolicies, error) {
	result := make(sidecarSvcAuthzClient.ScopedPolicies)

	if user == "" {
		return result, nil
	}

	allPolicies := a.policies.Copy()
	allGroups := a.roles.Copy()
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
			if sp, err := a.resolveGroupWithScope(g, binding.GetScope(), allPolicies); err != nil {
				return nil, err
			} else if sp != nil {
				result[groupName] = *sp
			}
		}
	}

	return result, nil
}

func (a *implementation) resolveGroupWithScope(g *sidecarSvcAuthzTypes.Role, scope *sidecarSvcAuthzTypes.Policy, allPolicies map[string]*sidecarSvcAuthzTypes.Policy) (*sidecarSvcAuthzClient.ScopedPolicy, error) {
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
