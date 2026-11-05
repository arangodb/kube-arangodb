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

// Test_ExtractGroups_GroupBindingUnion verifies that roles bound to a group are granted to a request
// only when the request's token carries that group (its JWT groups claim, passed as req.Roles), and are
// resolved from the dedicated group-binding pool.
func Test_ExtractGroups_GroupBindingUnion(t *testing.T) {
	allowRead := &sidecarSvcAuthzTypes.Policy{Statements: []*sidecarSvcAuthzTypes.PolicyStatement{
		{Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"database:read"}, Resources: []string{"*"}},
	}}
	scopeAll := &sidecarSvcAuthzTypes.Policy{Statements: []*sidecarSvcAuthzTypes.PolicyStatement{
		{Effect: sidecarSvcAuthzTypes.Effect_Allow, Actions: []string{"*"}, Resources: []string{"*"}},
	}}

	policies := map[string]*sidecarSvcAuthzTypes.Policy{"reader": allowRead}
	roles := map[string]*sidecarSvcAuthzTypes.Role{"editor": {Policies: []string{"reader"}}}
	// Group binding pool keyed group:role - "developers" grants the "editor" role with an open scope.
	groupBindings := map[string]*sidecarSvcAuthzTypes.UserRoleBinding{
		"developers:editor": {Role: "editor", Scope: scopeAll},
	}

	c := newCache(policies, roles, map[string]*sidecarSvcAuthzTypes.UserRoleBinding{}, groupBindings)

	req := func(groups ...string) *pbAuthorizationV1.AuthorizationV1PermissionRequest {
		return &pbAuthorizationV1.AuthorizationV1PermissionRequest{
			User:     util.NewType("alice"),
			Roles:    groups,
			Action:   "database:read",
			Resource: "mydb",
		}
	}

	t.Run("member of bound group is granted access", func(t *testing.T) {
		sp := c.extractGroups("alice", []string{"developers"})
		resp, err := sp.Evaluate(req("developers"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Allow, resp.GetEffect())
	})

	t.Run("member of an unrelated group is denied", func(t *testing.T) {
		sp := c.extractGroups("alice", []string{"other"})
		resp, err := sp.Evaluate(req("other"))
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})

	t.Run("no groups in token is denied", func(t *testing.T) {
		sp := c.extractGroups("alice", nil)
		resp, err := sp.Evaluate(req())
		require.NoError(t, err)
		require.Equal(t, sidecarSvcAuthzTypes.Effect_Deny, resp.GetEffect())
	})
}
