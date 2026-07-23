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

package reconcile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/permission"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
)

// Test_predefinedRoles_catalog verifies the predefined-role catalog: super-admin is the only
// role that ships with an Allow-all policy and a root binding, every role is namespaced under
// managed:predefined:, hyphenated and described, and names are unique.
func Test_predefinedRoles_catalog(t *testing.T) {
	require.Equal(t, "managed:predefined:", managedObjectPrefix)
	require.Equal(t, permission.ManagedPredefinedRoleName(permission.PredefinedRoleSuperAdmin), managedRoleName(permission.PredefinedRoleSuperAdmin))

	names := map[string]bool{}
	var bound []string

	for _, r := range predefinedRoles {
		require.NotEmpty(t, r.Description, "%s must have a description", r.Name)
		require.Regexp(t, `^[a-z]+(-[a-z]+)*$`, r.Name, "role name must be hyphenated lower-case")
		require.False(t, names[r.Name], "duplicate role %s", r.Name)
		names[r.Name] = true

		if r.BindRootUser {
			bound = append(bound, r.Name)
		}

		// Only super-admin carries a policy; the rest are empty containers.
		if r.Name == permission.PredefinedRoleSuperAdmin {
			require.NotEmpty(t, r.Statements, "super-admin must ship a policy")
			require.True(t, r.BindRootUser, "super-admin must be bound to root")
		} else {
			require.Empty(t, r.Statements, "%s must be created without a policy", r.Name)
			require.False(t, r.BindRootUser, "%s must not be bound to root", r.Name)
		}
	}

	require.Equal(t, []string{permission.PredefinedRoleSuperAdmin}, bound, "only super-admin is bound to the root user")

	// The catalog must cover exactly the exported PredefinedRoleNames (the single source of truth
	// shared with the handlers and the e2e tests) - no missing and no extra roles.
	catalog := make([]string, 0, len(predefinedRoles))
	for _, r := range predefinedRoles {
		catalog = append(catalog, r.Name)
	}
	require.ElementsMatch(t, permission.PredefinedRoleNames, catalog, "predefinedRoles must match permission.PredefinedRoleNames")

	// super-admin is the reserved role and must be reported as such.
	require.True(t, permission.IsReservedRoleName(permission.ManagedPredefinedRoleName(permission.PredefinedRoleSuperAdmin)))
	require.False(t, permission.IsReservedRoleName(permission.ManagedPredefinedRoleName(permission.PredefinedRoleCoreDBReader)))
}

// fakeAuthzClient is an in-memory authorization sidecar client. The embedded nil interface
// makes any method not overridden below panic, so the test fails loudly if the sync starts
// depending on an unexpected call.
type fakeAuthzClient struct {
	sidecarSvcAuthzDefinition.AuthorizationAPIClient

	policies map[string]*sidecarSvcAuthzTypes.Policy
	roles    map[string]*sidecarSvcAuthzTypes.Role
	bindings map[string]*sidecarSvcAuthzTypes.UserRoleBinding // key: user/role

	writes int
}

func newFakeAuthzClient() *fakeAuthzClient {
	return &fakeAuthzClient{
		policies: map[string]*sidecarSvcAuthzTypes.Policy{},
		roles:    map[string]*sidecarSvcAuthzTypes.Role{},
		bindings: map[string]*sidecarSvcAuthzTypes.UserRoleBinding{},
	}
}

func (f *fakeAuthzClient) APIGetPolicy(_ context.Context, in *sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest, _ ...grpc.CallOption) (*sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse, error) {
	if p, ok := f.policies[in.GetName()]; ok {
		return &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse{Item: p}, nil
	}
	return nil, grpcStatus.Error(codes.NotFound, "not found")
}

func (f *fakeAuthzClient) APICreatePolicy(_ context.Context, in *sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest, _ ...grpc.CallOption) (*sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse, error) {
	f.writes++
	f.policies[in.GetName()] = in.GetItem()
	return &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse{Item: in.GetItem()}, nil
}

func (f *fakeAuthzClient) APIUpdatePolicy(_ context.Context, in *sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest, _ ...grpc.CallOption) (*sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse, error) {
	f.writes++
	f.policies[in.GetName()] = in.GetItem()
	return &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyResponse{Item: in.GetItem()}, nil
}

func (f *fakeAuthzClient) APIGetRole(_ context.Context, in *sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest, _ ...grpc.CallOption) (*sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse, error) {
	if r, ok := f.roles[in.GetName()]; ok {
		return &sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse{Item: r}, nil
	}
	return nil, grpcStatus.Error(codes.NotFound, "not found")
}

func (f *fakeAuthzClient) APICreateRole(_ context.Context, in *sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest, _ ...grpc.CallOption) (*sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse, error) {
	f.writes++
	f.roles[in.GetName()] = in.GetItem()
	return &sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse{Item: in.GetItem()}, nil
}

func (f *fakeAuthzClient) APIUpdateRole(_ context.Context, in *sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest, _ ...grpc.CallOption) (*sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse, error) {
	f.writes++
	f.roles[in.GetName()] = in.GetItem()
	return &sidecarSvcAuthzDefinition.AuthorizationAPIRoleResponse{Item: in.GetItem()}, nil
}

func (f *fakeAuthzClient) APIListUserRoleBindings(_ context.Context, in *sidecarSvcAuthzDefinition.AuthorizationAPIUserRequest, _ ...grpc.CallOption) (*sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingListResponse, error) {
	var out []*sidecarSvcAuthzTypes.UserRoleBinding
	prefix := in.GetUser() + "/"
	for k, b := range f.bindings {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			out = append(out, b)
		}
	}
	return &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingListResponse{Bindings: out}, nil
}

func (f *fakeAuthzClient) APIAssignUserRole(_ context.Context, in *sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingRequest, _ ...grpc.CallOption) (*sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse, error) {
	f.writes++
	f.bindings[in.GetUser()+"/"+in.GetRole()] = &sidecarSvcAuthzTypes.UserRoleBinding{Role: in.GetRole(), Scope: in.GetScope()}
	return &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse{}, nil
}

func (f *fakeAuthzClient) APIReplaceUserRoleScope(_ context.Context, in *sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingRequest, _ ...grpc.CallOption) (*sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse, error) {
	f.writes++
	f.bindings[in.GetUser()+"/"+in.GetRole()] = &sidecarSvcAuthzTypes.UserRoleBinding{Role: in.GetRole(), Scope: in.GetScope()}
	return &sidecarSvcAuthzDefinition.AuthorizationAPIUserRoleBindingResponse{}, nil
}

// noBoundPolicies is a bound-policy resolver that returns nothing (no customer-attached policies).
func noBoundPolicies(string) ([]string, error) { return nil, nil }

func Test_syncRBACPermissions(t *testing.T) {
	ctx := context.Background()
	superRole := managedRoleName("super-admin")

	t.Run("creates the full catalog from empty", func(t *testing.T) {
		f := newFakeAuthzClient()

		require.NoError(t, syncRBACPermissions(ctx, f, noBoundPolicies))

		// Every predefined role exists, with its description.
		require.Len(t, f.roles, len(predefinedRoles))
		for _, r := range predefinedRoles {
			role, ok := f.roles[managedRoleName(r.Name)]
			require.True(t, ok, "role %s must be created", r.Name)
			require.Equal(t, r.Description, role.GetDescription())
		}

		// Only super-admin has a policy.
		require.Len(t, f.policies, 1)
		require.Contains(t, f.policies, managedRoleName("super-admin"))
		require.Equal(t, allowAllScope().Hash(), f.policies[managedRoleName("super-admin")].Hash())
		require.Equal(t, []string{managedRoleName("super-admin")}, f.roles[superRole].GetPolicies())

		// The empty roles carry no policy.
		require.Empty(t, f.roles[managedRoleName("coredb-reader")].GetPolicies())

		// Only the root user is bound, to super-admin, with an Allow-all scope.
		require.Len(t, f.bindings, 1)
		b, ok := f.bindings[api.UserNameRoot+"/"+superRole]
		require.True(t, ok)
		require.Equal(t, allowAllScope().Hash(), b.GetScope().Hash())
	})

	t.Run("idempotent - no writes when already in sync", func(t *testing.T) {
		f := newFakeAuthzClient()
		require.NoError(t, syncRBACPermissions(ctx, f, noBoundPolicies))

		f.writes = 0
		require.NoError(t, syncRBACPermissions(ctx, f, noBoundPolicies))
		require.Equal(t, 0, f.writes, "a second sync must not write anything")
	})

	t.Run("recreates a deleted role", func(t *testing.T) {
		f := newFakeAuthzClient()
		require.NoError(t, syncRBACPermissions(ctx, f, noBoundPolicies))

		delete(f.roles, managedRoleName("secret-admin"))

		f.writes = 0
		require.NoError(t, syncRBACPermissions(ctx, f, noBoundPolicies))
		require.Equal(t, 1, f.writes, "only the deleted role is recreated")
		require.Contains(t, f.roles, managedRoleName("secret-admin"))
	})

	t.Run("repairs a drifted super-admin policy", func(t *testing.T) {
		f := newFakeAuthzClient()
		require.NoError(t, syncRBACPermissions(ctx, f, noBoundPolicies))

		f.policies[managedRoleName("super-admin")] = &sidecarSvcAuthzTypes.Policy{
			Statements: []*sidecarSvcAuthzTypes.PolicyStatement{
				{Effect: sidecarSvcAuthzTypes.Effect_Deny, Actions: []string{"meta:GetKey"}, Resources: []string{"x"}},
			},
		}

		f.writes = 0
		require.NoError(t, syncRBACPermissions(ctx, f, noBoundPolicies))
		require.Equal(t, 1, f.writes, "only the drifted policy is rewritten")
		require.Equal(t, allowAllScope().Hash(), f.policies[managedRoleName("super-admin")].Hash())
	})

	t.Run("merges policies bound to a predefined role", func(t *testing.T) {
		f := newFakeAuthzClient()

		// A customer attaches a policy to coredb-reader via an ArangoPermissionPolicyRoleBinding.
		readerRole := managedRoleName("coredb-reader")
		bound := func(roleName string) ([]string, error) {
			if roleName == readerRole {
				return []string{"managed:operator:my-policy"}, nil
			}
			return nil, nil
		}

		require.NoError(t, syncRBACPermissions(ctx, f, bound))

		// The bound policy is attached to the predefined role...
		require.Equal(t, []string{"managed:operator:my-policy"}, f.roles[readerRole].GetPolicies())
		// ...and super-admin still keeps its own bundled policy.
		require.Equal(t, []string{superRole}, f.roles[superRole].GetPolicies())

		// Re-running without the binding removes the policy again (drift repair does not clobber
		// bound policies, but does remove ones that are no longer bound).
		f.writes = 0
		require.NoError(t, syncRBACPermissions(ctx, f, noBoundPolicies))
		require.Empty(t, f.roles[readerRole].GetPolicies())
	})
}
