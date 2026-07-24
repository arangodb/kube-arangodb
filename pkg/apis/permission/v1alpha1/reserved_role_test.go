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

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/apis/permission"
	permissionApiPolicy "github.com/arangodb/kube-arangodb/pkg/apis/permission/v1alpha1/policy"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

// Test_ReservedRole_NotAssignable asserts the operator refuses customer bindings that reference the
// reserved super-admin role - the enforcement behind the "not assignable by customers" guarantee.
func Test_ReservedRole_NotAssignable(t *testing.T) {
	depl := &sharedApi.Object{Name: "depl"}
	scope := &permissionApiPolicy.Policy{Statements: permissionApiPolicy.Statements{
		{Effect: permissionApiPolicy.EffectAllow, Actions: permissionApiPolicy.Actions{"*"}, Resources: permissionApiPolicy.Resources{"*"}},
	}}
	superAdmin := permission.ManagedPredefinedRoleName(permission.PredefinedRoleSuperAdmin)
	coreDBReader := permission.ManagedPredefinedRoleName(permission.PredefinedRoleCoreDBReader)

	t.Run("RoleUserBinding rejects super-admin", func(t *testing.T) {
		require.ErrorContains(t, (&ArangoPermissionRoleUserBindingSpec{
			Deployment: depl,
			Role:       &ArangoPermissionBindingRef{Direct: superAdmin},
			UserName:   "alice",
			Scope:      scope,
		}).Validate(), "reserved")
	})

	t.Run("RoleUserBinding allows a normal predefined role", func(t *testing.T) {
		require.NoError(t, (&ArangoPermissionRoleUserBindingSpec{
			Deployment: depl,
			Role:       &ArangoPermissionBindingRef{Direct: coreDBReader},
			UserName:   "alice",
			Scope:      scope,
		}).Validate())
	})

	t.Run("PolicyRoleBinding rejects super-admin", func(t *testing.T) {
		require.ErrorContains(t, (&ArangoPermissionPolicyRoleBindingSpec{
			Deployment: depl,
			Policy:     &ArangoPermissionBindingRef{Name: "p"},
			Role:       &ArangoPermissionBindingRef{Direct: superAdmin},
		}).Validate(), "reserved")
	})
}
