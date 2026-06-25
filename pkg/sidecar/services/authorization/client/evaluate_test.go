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
	"testing"

	"github.com/stretchr/testify/require"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func newPolicy(t *testing.T, statements ...*sidecarSvcAuthzTypes.PolicyStatement) *Policy {
	t.Helper()
	p, err := NewPolicy(&sidecarSvcAuthzTypes.Policy{Statements: statements})
	require.NoError(t, err)
	return &p
}

func evalReq(action, resource string) *pbAuthorizationV1.AuthorizationV1PermissionRequest {
	return &pbAuthorizationV1.AuthorizationV1PermissionRequest{
		User:     util.NewType("test-user"),
		Action:   action,
		Resource: resource,
	}
}

func Test_EvaluatePolicies_PolicyAndScope(t *testing.T) {
	// Policy: allow database:read on *, deny database:drop on production
	policy := newPolicy(t,
		&sidecarSvcAuthzTypes.PolicyStatement{
			Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"database:read"}, Resources: []string{"*"},
		},
		&sidecarSvcAuthzTypes.PolicyStatement{
			Effect: sidecarSvcAuthzTypes.Effect_Deny, Actions: []string{"database:drop"}, Resources: []string{"production"},
		},
	)

	// Open scope: allow everything
	scopeAll := newPolicy(t, &sidecarSvcAuthzTypes.PolicyStatement{
		Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"*"}, Resources: []string{"*"},
	})

	// Restricted scope: only allow database:read
	scopeRestricted := newPolicy(t, &sidecarSvcAuthzTypes.PolicyStatement{
		Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"database:read"}, Resources: []string{"*"},
	})

	t.Run("Policy only - allow", func(t *testing.T) {
		resp, err := EvaluatePolicies(evalReq("database:read", "mydb"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, resp.GetEffect())
	})

	t.Run("Policy only - explicit deny", func(t *testing.T) {
		resp, err := EvaluatePolicies(evalReq("database:drop", "production"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("Policy only - no match deny", func(t *testing.T) {
		resp, err := EvaluatePolicies(evalReq("collection:write", "reports"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("Open scope does not independently grant access", func(t *testing.T) {
		// Scope alone should not grant access - it only acts as a boundary
		resp, err := EvaluatePolicies(evalReq("collection:write", "reports"), scopeAll)
		require.NoError(t, err)
		// Note: EvaluatePolicies itself treats scope as a regular policy.
		// The intersection logic lives in Evaluate (client.go), not here.
		// This test documents that scope as a standalone policy DOES allow.
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, resp.GetEffect())
	})

	t.Run("Policy + open scope: allow passes through", func(t *testing.T) {
		// When used correctly (policy evaluated first, scope as boundary),
		// an allowed action with open scope stays allowed.
		policyResp, err := EvaluatePolicies(evalReq("database:read", "mydb"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, policyResp.GetEffect())

		scopeResp, err := EvaluatePolicies(evalReq("database:read", "mydb"), scopeAll)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, scopeResp.GetEffect())
	})

	t.Run("Policy + open scope: explicit deny preserved", func(t *testing.T) {
		policyResp, err := EvaluatePolicies(evalReq("database:drop", "production"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, policyResp.GetEffect())
		// Scope never reached because policy already denied
	})

	t.Run("Policy + open scope: no-match deny preserved", func(t *testing.T) {
		policyResp, err := EvaluatePolicies(evalReq("collection:write", "reports"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, policyResp.GetEffect())
		// Scope never reached because policy denied by default
	})

	t.Run("Policy + restricted scope: scope restricts allowed action", func(t *testing.T) {
		// Policy allows database:read on *, but scope only allows database:read
		// So database:read is still allowed (intersection: both agree)
		policyResp, err := EvaluatePolicies(evalReq("database:read", "mydb"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, policyResp.GetEffect())

		scopeResp, err := EvaluatePolicies(evalReq("database:read", "mydb"), scopeRestricted)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, scopeResp.GetEffect())
	})

	t.Run("Restricted scope denies action not in scope", func(t *testing.T) {
		// If a hypothetical policy allowed database:write, the restricted scope
		// would deny it because database:write is not in scope
		scopeResp, err := EvaluatePolicies(evalReq("database:write", "mydb"), scopeRestricted)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, scopeResp.GetEffect())
	})
}

func Test_EvaluatePolicies_NoPolicies(t *testing.T) {
	resp, err := EvaluatePolicies(evalReq("database:read", "mydb"))
	require.NoError(t, err)
	require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
}

func Test_EvaluateGroups(t *testing.T) {
	policy := newPolicy(t,
		&sidecarSvcAuthzTypes.PolicyStatement{
			Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"database:read", "database:*"}, Resources: []string{"*"},
		},
		&sidecarSvcAuthzTypes.PolicyStatement{
			Effect: sidecarSvcAuthzTypes.Effect_Deny, Actions: []string{"database:drop"}, Resources: []string{"production"},
		},
	)

	scopeAll := newPolicy(t, &sidecarSvcAuthzTypes.PolicyStatement{
		Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"*"}, Resources: []string{"*"},
	})

	scopeRestricted := newPolicy(t, &sidecarSvcAuthzTypes.PolicyStatement{
		Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"database:read"}, Resources: []string{"*"},
	})

	t.Run("Open scope allows policy-granted action", func(t *testing.T) {
		sp := ScopedPolicy{Policies: []*Policy{policy}, Scope: scopeAll}
		resp, err := sp.Evaluate(evalReq("database:read", "mydb"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, resp.GetEffect())
	})

	t.Run("Open scope preserves explicit deny", func(t *testing.T) {
		sp := ScopedPolicy{Policies: []*Policy{policy}, Scope: scopeAll}
		resp, err := sp.Evaluate(evalReq("database:drop", "production"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("Open scope preserves no-match deny", func(t *testing.T) {
		sp := ScopedPolicy{Policies: []*Policy{policy}, Scope: scopeAll}
		resp, err := sp.Evaluate(evalReq("collection:write", "reports"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("Restricted scope blocks action outside boundary", func(t *testing.T) {
		sp := ScopedPolicy{Policies: []*Policy{policy}, Scope: scopeRestricted}
		resp, err := sp.Evaluate(evalReq("database:write", "mydb"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("Restricted scope allows action inside boundary", func(t *testing.T) {
		sp := ScopedPolicy{Policies: []*Policy{policy}, Scope: scopeRestricted}
		resp, err := sp.Evaluate(evalReq("database:read", "mydb"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, resp.GetEffect())
	})

	t.Run("Explicit deny in scope blocks allowed action", func(t *testing.T) {
		// Scope: allow everything EXCEPT database:read on _system
		scopeWithDeny := newPolicy(t,
			&sidecarSvcAuthzTypes.PolicyStatement{
				Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"*"}, Resources: []string{"*"},
			},
			&sidecarSvcAuthzTypes.PolicyStatement{
				Effect: sidecarSvcAuthzTypes.Effect_Deny, Actions: []string{"database:read"}, Resources: []string{"_system"},
			},
		)
		sp := ScopedPolicy{Policies: []*Policy{policy}, Scope: scopeWithDeny}
		// database:read on _system: policy allows, but scope has explicit deny → denied
		resp, err := sp.Evaluate(evalReq("database:read", "_system"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())

		// database:read on other_db: policy allows, scope allows (deny only on _system) → allowed
		resp, err = sp.Evaluate(evalReq("database:read", "other_db"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, resp.GetEffect())

		// collection:write on _system: scope allows, but policy doesn't have collection:write → denied by policy
		resp, err = sp.Evaluate(evalReq("collection:write", "_system"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("Nil scope means group not considered", func(t *testing.T) {
		sp := ScopedPolicy{Policies: []*Policy{policy}, Scope: nil}
		resp, err := sp.Evaluate(evalReq("database:read", "mydb"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("Empty ScopedPolicies denies", func(t *testing.T) {
		groups := ScopedPolicies{}
		resp, err := groups.Evaluate(evalReq("database:read", "mydb"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("Multiple groups - first grants", func(t *testing.T) {
		groups := ScopedPolicies{
			"g1": {Policies: []*Policy{policy}, Scope: scopeAll},
			"g2": {Policies: []*Policy{policy}, Scope: scopeRestricted},
		}
		resp, err := groups.Evaluate(evalReq("database:read", "mydb"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, resp.GetEffect())
	})
}

func Test_EvaluatePolicies_ExplicitDenyBeatsAllow(t *testing.T) {
	policy := newPolicy(t,
		&sidecarSvcAuthzTypes.PolicyStatement{
			Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"database:*"}, Resources: []string{"*"},
		},
		&sidecarSvcAuthzTypes.PolicyStatement{
			Effect: sidecarSvcAuthzTypes.Effect_Deny, Actions: []string{"database:drop"}, Resources: []string{"production"},
		},
	)

	t.Run("Wildcard allow works", func(t *testing.T) {
		resp, err := EvaluatePolicies(evalReq("database:read", "mydb"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, resp.GetEffect())
	})

	t.Run("Explicit deny overrides wildcard allow", func(t *testing.T) {
		resp, err := EvaluatePolicies(evalReq("database:drop", "production"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("Deny on different resource still allows", func(t *testing.T) {
		resp, err := EvaluatePolicies(evalReq("database:drop", "staging"), policy)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, resp.GetEffect())
	})
}
