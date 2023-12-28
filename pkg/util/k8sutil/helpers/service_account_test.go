//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	rbac "k8s.io/api/rbac/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_ServiceAccount_Roles(t *testing.T) {
	k := fake.NewSimpleClientset()

	var obj sharedApi.ServiceAccount

	t.Run("PreCheck", func(t *testing.T) {
		require.Nil(t, obj.Object)
	})

	t.Run("Create SA without any roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, nil, nil)
		}))

		require.NotNil(t, obj.Object)
		require.Nil(t, obj.Namespaced)
		require.Nil(t, obj.Cluster)
	})

	t.Run("Create SA with roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, []rbac.PolicyRule{
				{
					Resources: []string{"*"},
				},
			}, nil)
		}))

		require.NotNil(t, obj.Object)
		require.NotNil(t, obj.Namespaced)
		require.NotNil(t, obj.Namespaced.Binding)
		require.NotNil(t, obj.Namespaced.Role)
		require.Nil(t, obj.Cluster)

		sa := tests.NewMetaObject[*rbac.Role](t, tests.FakeNamespace, obj.Namespaced.Role.GetName())

		tests.RefreshObjects(t, k, nil, &sa)

		require.Len(t, sa.Rules, 1)
		require.Len(t, sa.Rules[0].Resources, 1)
		require.Equal(t, sa.Rules[0].Resources[0], "*")
	})

	t.Run("Create SA with updated roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, []rbac.PolicyRule{
				{
					Resources: []string{"DATA"},
				},
			}, nil)
		}))

		require.NotNil(t, obj.Object)
		require.NotNil(t, obj.Namespaced)
		require.NotNil(t, obj.Namespaced.Binding)
		require.NotNil(t, obj.Namespaced.Role)
		require.Nil(t, obj.Cluster)

		sa := tests.NewMetaObject[*rbac.Role](t, tests.FakeNamespace, obj.Namespaced.Role.GetName())

		tests.RefreshObjects(t, k, nil, &sa)

		require.Len(t, sa.Rules, 1)
		require.Len(t, sa.Rules[0].Resources, 1)
		require.Equal(t, sa.Rules[0].Resources[0], "DATA")
	})

	t.Run("Create SA with multiple roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, []rbac.PolicyRule{
				{
					Resources: []string{"DATA"},
				},
				{
					Resources: []string{"*"},
				},
			}, nil)
		}))

		require.NotNil(t, obj.Object)
		require.NotNil(t, obj.Namespaced)
		require.NotNil(t, obj.Namespaced.Binding)
		require.NotNil(t, obj.Namespaced.Role)
		require.Nil(t, obj.Cluster)

		sa := tests.NewMetaObject[*rbac.Role](t, tests.FakeNamespace, obj.Namespaced.Role.GetName())

		tests.RefreshObjects(t, k, nil, &sa)

		require.Len(t, sa.Rules, 2)
		require.Len(t, sa.Rules[0].Resources, 1)
		require.Equal(t, sa.Rules[0].Resources[0], "DATA")
	})

	t.Run("Create SA with updated multiple roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, []rbac.PolicyRule{
				{
					Resources: []string{"*"},
				},
				{
					Resources: []string{"DATA"},
				},
			}, nil)
		}))

		require.NotNil(t, obj.Object)
		require.NotNil(t, obj.Namespaced)
		require.NotNil(t, obj.Namespaced.Binding)
		require.NotNil(t, obj.Namespaced.Role)
		require.Nil(t, obj.Cluster)

		sa := tests.NewMetaObject[*rbac.Role](t, tests.FakeNamespace, obj.Namespaced.Role.GetName())

		tests.RefreshObjects(t, k, nil, &sa)

		require.Len(t, sa.Rules, 2)
		require.Len(t, sa.Rules[0].Resources, 1)
		require.Equal(t, sa.Rules[0].Resources[0], "*")
	})

	t.Run("Remove SA Roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, nil, nil)
		}))

		require.NotNil(t, obj.Object)
		require.Nil(t, obj.Namespaced)
		require.Nil(t, obj.Cluster)
	})
}

func Test_ServiceAccount_ClusterRoles(t *testing.T) {
	k := fake.NewSimpleClientset()

	var obj sharedApi.ServiceAccount

	t.Run("PreCheck", func(t *testing.T) {
		require.Nil(t, obj.Object)
	})

	t.Run("Create SA without any roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, nil, nil)
		}))

		require.NotNil(t, obj.Object)
		require.Nil(t, obj.Cluster)
		require.Nil(t, obj.Namespaced)
	})

	t.Run("Create SA with roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, nil, []rbac.PolicyRule{
				{
					Resources: []string{"*"},
				},
			})
		}))

		require.NotNil(t, obj.Object)
		require.NotNil(t, obj.Cluster)
		require.NotNil(t, obj.Cluster.Binding)
		require.NotNil(t, obj.Cluster.Role)
		require.Nil(t, obj.Namespaced)

		sa := tests.NewMetaObject[*rbac.ClusterRole](t, tests.FakeNamespace, obj.Cluster.Role.GetName())

		tests.RefreshObjects(t, k, nil, &sa)

		require.Len(t, sa.Rules, 1)
		require.Len(t, sa.Rules[0].Resources, 1)
		require.Equal(t, sa.Rules[0].Resources[0], "*")
	})

	t.Run("Create SA with updated roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, nil, []rbac.PolicyRule{
				{
					Resources: []string{"DATA"},
				},
			})
		}))

		require.NotNil(t, obj.Object)
		require.NotNil(t, obj.Cluster)
		require.NotNil(t, obj.Cluster.Binding)
		require.NotNil(t, obj.Cluster.Role)
		require.Nil(t, obj.Namespaced)

		sa := tests.NewMetaObject[*rbac.ClusterRole](t, tests.FakeNamespace, obj.Cluster.Role.GetName())

		tests.RefreshObjects(t, k, nil, &sa)

		require.Len(t, sa.Rules, 1)
		require.Len(t, sa.Rules[0].Resources, 1)
		require.Equal(t, sa.Rules[0].Resources[0], "DATA")
	})

	t.Run("Create SA with multiple roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, nil, []rbac.PolicyRule{
				{
					Resources: []string{"DATA"},
				},
				{
					Resources: []string{"*"},
				},
			})
		}))

		require.NotNil(t, obj.Object)
		require.NotNil(t, obj.Cluster)
		require.NotNil(t, obj.Cluster.Binding)
		require.NotNil(t, obj.Cluster.Role)
		require.Nil(t, obj.Namespaced)

		sa := tests.NewMetaObject[*rbac.ClusterRole](t, tests.FakeNamespace, obj.Cluster.Role.GetName())

		tests.RefreshObjects(t, k, nil, &sa)

		require.Len(t, sa.Rules, 2)
		require.Len(t, sa.Rules[0].Resources, 1)
		require.Equal(t, sa.Rules[0].Resources[0], "DATA")
	})

	t.Run("Create SA with updated multiple roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, nil, []rbac.PolicyRule{
				{
					Resources: []string{"*"},
				},
				{
					Resources: []string{"DATA"},
				},
			})
		}))

		require.NotNil(t, obj.Object)
		require.NotNil(t, obj.Cluster)
		require.NotNil(t, obj.Cluster.Binding)
		require.NotNil(t, obj.Cluster.Role)
		require.Nil(t, obj.Namespaced)

		sa := tests.NewMetaObject[*rbac.ClusterRole](t, tests.FakeNamespace, obj.Cluster.Role.GetName())

		tests.RefreshObjects(t, k, nil, &sa)

		require.Len(t, sa.Rules, 2)
		require.Len(t, sa.Rules[0].Resources, 1)
		require.Equal(t, sa.Rules[0].Resources[0], "*")
	})

	t.Run("Remove SA Roles", func(t *testing.T) {
		require.NoError(t, tests.HandleFunc(func(ctx context.Context) (bool, error) {
			return EnsureServiceAccount(ctx, k, meta.OwnerReference{}, &obj, "test", tests.FakeNamespace, nil, nil)
		}))

		require.NotNil(t, obj.Object)
		require.Nil(t, obj.Cluster)
		require.Nil(t, obj.Namespaced)
	})
}
