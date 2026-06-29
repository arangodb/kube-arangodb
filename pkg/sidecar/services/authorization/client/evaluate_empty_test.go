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

	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
)

// Test_Policy_EmptyAction_NotGranted ensures an empty request action is never
// granted, even against a wildcard policy/scope. A `*` matcher otherwise matches
// the empty string and would incorrectly grant access.
func Test_Policy_EmptyAction_NotGranted(t *testing.T) {
	allowAll := newPolicy(t, &sidecarSvcAuthzTypes.PolicyStatement{
		Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"*"}, Resources: []string{"*"},
	})

	t.Run("Policy.Evaluate: empty action does not match wildcard", func(t *testing.T) {
		allowed, err := allowAll.Evaluate("", "mydb", nil)
		require.NoError(t, err)
		require.False(t, allowed)
	})

	t.Run("Policy.Evaluate: empty action and empty resource", func(t *testing.T) {
		allowed, err := allowAll.Evaluate("", "", nil)
		require.NoError(t, err)
		require.False(t, allowed)
	})

	t.Run("Policy.Evaluate: non-empty action still matches wildcard", func(t *testing.T) {
		allowed, err := allowAll.Evaluate("database:read", "mydb", nil)
		require.NoError(t, err)
		require.True(t, allowed)
	})

	t.Run("EvaluatePolicies: empty action denied against wildcard", func(t *testing.T) {
		resp, err := EvaluatePolicies(evalReq("", "mydb"), allowAll)
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("ScopedPolicy: empty action denied even with open scope and policy", func(t *testing.T) {
		sp := ScopedPolicy{Policies: []*Policy{allowAll}, Scope: allowAll}
		resp, err := sp.Evaluate(evalReq("", "mydb"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})
}

// Test_Policy_EmptyActionsList_GrantsNothing documents that a statement with an
// empty actions list is inert - it matches no action and grants nothing. This is
// why an empty actions list is allowed rather than rejected at validation.
func Test_Policy_EmptyActionsList_GrantsNothing(t *testing.T) {
	emptyActions := newPolicy(t, &sidecarSvcAuthzTypes.PolicyStatement{
		Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{}, Resources: []string{"*"},
	})

	allowed, err := emptyActions.Evaluate("database:read", "mydb", nil)
	require.NoError(t, err)
	require.False(t, allowed)

	resp, err := EvaluatePolicies(evalReq("database:read", "mydb"), emptyActions)
	require.NoError(t, err)
	require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
}
