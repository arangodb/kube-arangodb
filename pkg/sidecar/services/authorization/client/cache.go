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
	"sort"

	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type cachedRole struct {
	policies []string
	scope    *Policy
}

func newCache(policies map[string]*sidecarSvcAuthzTypes.Policy, roles map[string]*sidecarSvcAuthzTypes.Role, userRoleBindings map[string]*sidecarSvcAuthzTypes.UserRoleBinding) internalCache {
	parsedPolicies := make(map[string]*Policy, len(policies))

	for name, policy := range policies {
		p, err := NewPolicy(policy)
		if err != nil {
			logger.Err(err).Str("name", name).Warn("Failed to create policy")
			continue
		}

		parsedPolicies[name] = &p
	}

	parsedRoles := make(map[string]cachedRole)

	for name, role := range roles {
		if role == nil {
			continue
		}

		cr := cachedRole{}

		cr.policies = append(cr.policies, role.Policies...)

		cr.policies = util.UniqueList(cr.policies)
		sort.Strings(cr.policies)

		if scope := role.GetScope(); scope != nil {
			if p, err := NewPolicy(scope); err != nil {
				logger.Err(err).Str("role", name).Warn("Failed to parse role scope policy")
			} else {
				cr.scope = &p
			}
		}

		parsedRoles[name] = cr
	}

	return internalCache{
		roles:            parsedRoles,
		policies:         parsedPolicies,
		userRoleBindings: userRoleBindings,
	}
}

type internalCache struct {
	roles            map[string]cachedRole
	policies         map[string]*Policy
	userRoleBindings map[string]*sidecarSvcAuthzTypes.UserRoleBinding
}

func (c *internalCache) extractGroups(user string, groupNames ...string) ScopedPolicies {
	if c == nil {
		return nil
	}

	result := make(ScopedPolicies)

	// Resolve explicit group names
	for _, name := range groupNames {
		c.resolveGroup(name, result)
	}

	// Resolve groups from user bindings
	if user != "" {
		prefix := user + ":"
		for key, binding := range c.userRoleBindings {
			if len(key) <= len(prefix) || key[:len(prefix)] != prefix {
				continue
			}
			c.resolveGroup(binding.GetRole(), result)
		}
	}

	return result
}

func (c *internalCache) resolveGroup(name string, result ScopedPolicies) {
	if _, exists := result[name]; exists {
		return
	}

	g, ok := c.roles[name]
	if !ok || g.scope == nil {
		return
	}

	sp := ScopedPolicy{Scope: g.scope}

	for _, policyName := range g.policies {
		if p, ok := c.policies[policyName]; ok {
			sp.Policies = append(sp.Policies, p)
		}
	}

	result[name] = sp
}
