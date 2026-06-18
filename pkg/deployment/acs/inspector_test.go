//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package acs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func inspect(t *testing.T, depl *api.ArangoDeployment, c kclient.Client) error {
	i := tests.NewInspector(t, c)
	a := NewACS("", i)
	return a.Inspect(context.Background(), depl, c, i)
}
func Test_InspectACS(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "test")
		c := kclient.NewFakeClientBuilder().Add(deployment).Client()
		require.NoError(t, inspect(t, deployment, c))
	})
	t.Run("Valid deployment", func(t *testing.T) {
		deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "test")
		acs := tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test")
		acs.Spec.DeploymentName = deployment.GetName()
		c := kclient.NewFakeClientBuilder().Add(deployment, acs).Client()
		require.NoError(t, inspect(t, deployment, c))
		tests.RefreshObjects(t, c.Kubernetes(), c.Arango(), &acs)
		require.NotNil(t, acs.Status.Deployment)
		require.Equal(t, acs.Status.Deployment.Name, deployment.GetName())
		require.Equal(t, acs.Status.Deployment.Namespace, deployment.GetNamespace())
		require.Equal(t, acs.Status.Deployment.UID, deployment.GetUID())
		require.True(t, acs.Status.Conditions.IsTrue(sutil.DeploymentReadyCondition))
	})
	t.Run("Valid deployment with multi acs", func(t *testing.T) {
		deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "test")
		acss := []*api.ArangoClusterSynchronization{
			tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test"),
			tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test1"),
			tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test2"),
			tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test3"),
		}
		acss[0].Spec.DeploymentName = deployment.GetName()
		acss[1].Spec.DeploymentName = deployment.GetName()
		acss[2].Spec.DeploymentName = deployment.GetName()
		acss[3].Spec.DeploymentName = deployment.GetName()
		f := kclient.NewFakeClientBuilder().Add(deployment)
		for _, o := range acss {
			f = f.Add(o)
		}
		c := f.Client()
		require.NoError(t, inspect(t, deployment, c))
		for id := range acss {
			tests.RefreshObjects(t, c.Kubernetes(), c.Arango(), &acss[id])
		}
		for _, acs := range acss {
			require.NotNil(t, acs.Status.Deployment)
			require.Equal(t, acs.Status.Deployment.Name, deployment.GetName())
			require.Equal(t, acs.Status.Deployment.Namespace, deployment.GetNamespace())
			require.Equal(t, acs.Status.Deployment.UID, deployment.GetUID())
			require.True(t, acs.Status.Conditions.IsTrue(sutil.DeploymentReadyCondition))
		}
	})
	t.Run("Valid deployment with multi acs - filter", func(t *testing.T) {
		deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "test")
		acss := []*api.ArangoClusterSynchronization{
			tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test"),
			tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test1"),
			tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test2"),
			tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test3"),
		}
		acss[1].Spec.DeploymentName = deployment.GetName()
		acss[3].Spec.DeploymentName = deployment.GetName()
		f := kclient.NewFakeClientBuilder().Add(deployment)
		for _, o := range acss {
			f = f.Add(o)
		}
		c := f.Client()
		require.NoError(t, inspect(t, deployment, c))
		i := tests.NewInspector(t, c)
		a, err := i.ArangoClusterSynchronization().V1()
		require.NoError(t, err)
		acss = a.Filter(arangoClusterSynchronizationFilter(deployment))
		require.Len(t, acss, 2)
	})
	t.Run("Recreated deployment", func(t *testing.T) {
		deployment := tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "test")
		acs := tests.NewMetaObject[*api.ArangoClusterSynchronization](t, tests.FakeNamespace, "test")
		acs.Spec.DeploymentName = deployment.GetName()
		c := kclient.NewFakeClientBuilder().Add(deployment, acs).Client()
		require.NoError(t, inspect(t, deployment, c))
		tests.RefreshObjects(t, c.Kubernetes(), c.Arango(), &acs)
		require.NotNil(t, acs.Status.Deployment)
		require.Equal(t, acs.Status.Deployment.Name, deployment.GetName())
		require.Equal(t, acs.Status.Deployment.Namespace, deployment.GetNamespace())
		require.Equal(t, acs.Status.Deployment.UID, deployment.GetUID())
		require.True(t, acs.Status.Conditions.IsTrue(sutil.DeploymentReadyCondition))
		// Recreate
		deployment = tests.NewMetaObject[*api.ArangoDeployment](t, tests.FakeNamespace, "test")
		err := c.Arango().DatabaseV1().ArangoDeployments(deployment.GetNamespace()).Delete(context.Background(), deployment.GetName(), meta.DeleteOptions{})
		require.NoError(t, err)
		_, err = c.Arango().DatabaseV1().ArangoDeployments(deployment.GetNamespace()).Create(context.Background(), deployment, meta.CreateOptions{})
		require.NoError(t, err)
		require.NoError(t, inspect(t, deployment, c))
		tests.RefreshObjects(t, c.Kubernetes(), c.Arango(), &acs)
		require.NotNil(t, acs.Status.Deployment)
		require.Equal(t, acs.Status.Deployment.Name, deployment.GetName())
		require.Equal(t, acs.Status.Deployment.Namespace, deployment.GetNamespace())
		require.NotEqual(t, acs.Status.Deployment.UID, deployment.GetUID())
		require.False(t, acs.Status.Conditions.IsTrue(sutil.DeploymentReadyCondition))
	})
}
