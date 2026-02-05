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

func newCache(policies map[string]*sidecarSvcAuthzTypes.Policy, roles map[string]*sidecarSvcAuthzTypes.Role) cache {
	parsedPolicies := make(map[string]*policy, len(policies))

	for name, policy := range policies {
		p, err := newPolicy(policy)
		if err != nil {
			logger.Err(err).Str("name", name).Warn("Failed to create policy")
			continue
		}

		parsedPolicies[name] = &p
	}

	parsedUsers := make(map[string][]string)
	parsedRoles := make(map[string][]string)

	for name, role := range roles {
		for _, policy := range role.Policies {
			parsedRoles[name] = append(parsedRoles[name], policy)

			for _, user := range role.Users {
				parsedUsers[user] = append(parsedUsers[user], policy)
			}
		}
	}

	for k := range parsedRoles {
		v := util.UniqueList(parsedRoles[k])
		sort.Strings(v)
		parsedRoles[k] = v
	}

	for k := range parsedUsers {
		v := util.UniqueList(parsedUsers[k])
		sort.Strings(v)
		parsedUsers[k] = v
	}

	return cache{
		users:    parsedUsers,
		roles:    parsedRoles,
		policies: parsedPolicies,
	}
}

type cache struct {
	users    map[string][]string
	roles    map[string][]string
	policies map[string]*policy
}

func (c *cache) extractUserPolicies(user string, userRoles ...string) []*policy {
	if c == nil {
		return nil
	}

	var policyNames = make(map[string]*policy, len(c.policies))

	for _, name := range c.users[user] {
		if p, ok := c.policies[name]; ok {
			policyNames[name] = p
		}
	}

	for _, roleName := range userRoles {
		if role, ok := c.roles[roleName]; ok {
			for _, name := range role {
				if p, ok := c.policies[name]; ok {
					policyNames[name] = p
				}
			}
		}
	}

	return util.MapValues(policyNames)
}
