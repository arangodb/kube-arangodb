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
	"testing"

	"github.com/stretchr/testify/require"

	pbImplAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	sidecarSvcAuthzDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/definition"
	"github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/pool"
	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

// fakePooler is a minimal in-memory pool.Pooler used to assert that the API layer refreshes
// the pool on get/update/list operations. It records how many times Refresh was called.
type fakePooler[T pool.PoolerObject] struct {
	items    map[string]T
	refreshN int
}

func newFakePooler[T pool.PoolerObject]() *fakePooler[T] {
	return &fakePooler[T]{items: map[string]T{}}
}

func (f *fakePooler[T]) Refresh(ctx context.Context) error { f.refreshN++; return nil }

func (f *fakePooler[T]) Create(ctx context.Context, name string, obj T) (T, uint32, error) {
	f.items[name] = obj
	return obj, 0, nil
}

func (f *fakePooler[T]) Update(ctx context.Context, name string, obj T) (T, uint32, error) {
	f.items[name] = obj
	return obj, 0, nil
}

func (f *fakePooler[T]) Delete(ctx context.Context, name string) (uint32, error) {
	delete(f.items, name)
	return 0, nil
}

func (f *fakePooler[T]) Item(name string) (T, uint32, bool) {
	v, ok := f.items[name]
	return v, 0, ok
}

func (f *fakePooler[T]) Index() uint32 { return 0 }

func (f *fakePooler[T]) Ready() bool { return true }

func (f *fakePooler[T]) Pool(start uint32) ([]pool.OffsetItem[T], error) { return nil, nil }

func (f *fakePooler[T]) Offsets() []pool.OffsetItem[T] { return nil }

func (f *fakePooler[T]) Items() []string {
	names := make([]string, 0, len(f.items))
	for name := range f.items {
		names = append(names, name)
	}
	return names
}

func (f *fakePooler[T]) Copy() map[string]T { return f.items }

func newRefreshTestImpl() (*implementation, *fakePooler[*sidecarSvcAuthzTypes.Policy], *fakePooler[*sidecarSvcAuthzTypes.Role], *fakePooler[*sidecarSvcAuthzTypes.UserRoleBinding], *fakePooler[*sidecarSvcAuthzTypes.UserRoleBinding]) {
	policies := newFakePooler[*sidecarSvcAuthzTypes.Policy]()
	roles := newFakePooler[*sidecarSvcAuthzTypes.Role]()
	userRoleBindings := newFakePooler[*sidecarSvcAuthzTypes.UserRoleBinding]()
	groupRoleBindings := newFakePooler[*sidecarSvcAuthzTypes.UserRoleBinding]()

	a := &implementation{
		policies:          policies,
		roles:             roles,
		userRoleBindings:  userRoleBindings,
		groupRoleBindings: groupRoleBindings,
		// Always => permissive plugin, so permission checks pass without an identity.
		authType: pbImplAuthorizationV1.ConfigurationTypeAlways,
	}

	return a, policies, roles, userRoleBindings, groupRoleBindings
}

// Test_API_RefreshesPools asserts that the RBAC API refreshes the backing pool from the
// store on every get/update/list operation (rather than serving stale in-memory state).
func Test_API_RefreshesPools(t *testing.T) {
	// Carry an authenticated identity as the gRPC auth interceptor would; combined with the
	// Always (permissive) plugin the permission checks pass.
	ctx := authenticator.WithIdentity(context.Background(), &authenticator.Identity{User: util.NewType("test-user")})

	t.Run("Policy List/Get/Update", func(t *testing.T) {
		a, policies, _, _, _ := newRefreshTestImpl()
		policies.items["p1"] = &sidecarSvcAuthzTypes.Policy{}

		_, err := a.APIListPolicy(ctx, &pbSharedV1.OffsetRequest{})
		require.NoError(t, err)
		require.Equal(t, 1, policies.refreshN)

		_, err = a.APIGetPolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{Name: "p1"})
		require.NoError(t, err)
		require.Equal(t, 2, policies.refreshN)

		_, err = a.APIUpdatePolicy(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIPolicyRequest{
			Name: "p1",
			Item: &sidecarSvcAuthzTypes.Policy{},
		})
		require.NoError(t, err)
		require.Equal(t, 3, policies.refreshN)
	})

	t.Run("Role List/Get/Update", func(t *testing.T) {
		a, _, roles, _, _ := newRefreshTestImpl()
		roles.items["r1"] = &sidecarSvcAuthzTypes.Role{}

		_, err := a.APIListRole(ctx, &pbSharedV1.OffsetRequest{})
		require.NoError(t, err)
		require.Equal(t, 1, roles.refreshN)

		_, err = a.APIGetRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPINamedRequest{Name: "r1"})
		require.NoError(t, err)
		require.Equal(t, 2, roles.refreshN)

		_, err = a.APIUpdateRole(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIRoleRequest{
			Name: "r1",
			Item: &sidecarSvcAuthzTypes.Role{},
		})
		require.NoError(t, err)
		require.Equal(t, 3, roles.refreshN)
	})

	t.Run("UserRoleBinding List", func(t *testing.T) {
		a, _, _, userRoleBindings, _ := newRefreshTestImpl()

		_, err := a.APIListUserRoleBindings(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIUserRequest{User: "alice"})
		require.NoError(t, err)
		require.Equal(t, 1, userRoleBindings.refreshN)
	})

	t.Run("GroupRoleBinding List", func(t *testing.T) {
		a, _, _, _, groupRoleBindings := newRefreshTestImpl()

		_, err := a.APIListGroupRoleBindings(ctx, &sidecarSvcAuthzDefinition.AuthorizationAPIGroupRequest{Group: "admins"})
		require.NoError(t, err)
		require.Equal(t, 1, groupRoleBindings.refreshN)
	})
}
