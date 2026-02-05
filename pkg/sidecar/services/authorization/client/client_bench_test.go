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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
)

func generateTestingPolicies(size, statements int) map[string]*sidecarSvcAuthzTypes.Policy {
	res := make(map[string]*sidecarSvcAuthzTypes.Policy, size)
	for id := 0; id < size; id++ {
		name := fmt.Sprintf("policy-%09d", id)

		var policy sidecarSvcAuthzTypes.Policy

		policy.Statements = make([]*sidecarSvcAuthzTypes.PolicyStatement, statements)
		for sid := 0; sid < statements; sid++ {
			var statement sidecarSvcAuthzTypes.PolicyStatement

			statement.Actions = []string{"*"}
			statement.Resources = []string{"*"}
			statement.Effect = sidecarSvcAuthzTypes.Effect_Allow

			policy.Statements[sid] = &statement
		}

		res[name] = &policy
	}

	return res
}

func generateTestingRoles(size int) map[string]*sidecarSvcAuthzTypes.Role {
	res := make(map[string]*sidecarSvcAuthzTypes.Role, size)
	for id := 0; id < size; id++ {
		name := fmt.Sprintf("role-%09d", id)

		var role sidecarSvcAuthzTypes.Role

		role.Policies = []string{
			fmt.Sprintf("policy-%09d", id),
		}

		res[name] = &role
	}

	return res
}

func BenchmarkClientEvaluationPerformance(b *testing.B) {
	var c client

	c.setPolicies(generateTestingPolicies(1024, 128))
	c.setRoles(generateTestingRoles(1024))

	c.policies.updated = time.Now().Add(time.Hour)
	c.roles.updated = time.Now().Add(time.Hour)

	var p pbImplAuthorizationV1Shared.Plugin = &c

	p = pbImplAuthorizationV1Shared.CachedPlugin(p)

	b.ResetTimer()

	for b.Loop() {
		res, err := p.Evaluate(b.Context(), &pbAuthorizationV1.AuthorizationV1PermissionRequest{
			User: "test",
			Roles: []string{
				fmt.Sprintf("role-%09d", rand.Intn(100)),
			},
			Action:   "test:GetData",
			Resource: "storage:/data/super/get",
		})
		require.NoError(b, err)
		require.EqualValues(b, pbAuthorizationV1.AuthorizationV1Effect_Allow, res.Effect)
	}
}
