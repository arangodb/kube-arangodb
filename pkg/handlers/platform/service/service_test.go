//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/suite"
)

func Test_ServiceReconcile(t *testing.T) {
	handler, ns, chartHandler := newFakeHandler(t)

	// Arrange
	extension := tests.NewMetaObject[*platformApi.ArangoPlatformService](t, ns, "example",
		func(t *testing.T, obj *platformApi.ArangoPlatformService) {})

	chart := tests.NewMetaObject[*platformApi.ArangoPlatformChart](t, ns, "secret",
		func(t *testing.T, obj *platformApi.ArangoPlatformChart) {})

	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, ns, "example",
		func(t *testing.T, obj *api.ArangoDeployment) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension, &chart, &deployment)

	t.Run("Missing chart section", func(t *testing.T) {
		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
	})

	t.Run("Missing deployment", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *platformApi.ArangoPlatformService) {
			obj.Spec.Chart = &sharedApi.Object{
				Name: "unknown",
			}
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "unknown",
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
	})

	t.Run("Existing deployment", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *platformApi.ArangoPlatformService) {
			obj.Spec.Deployment = &sharedApi.Object{
				Name: "example",
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)
	})

	t.Run("Missing chart", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *platformApi.ArangoPlatformService) {
			obj.Spec.Chart = &sharedApi.Object{
				Name: "unknown",
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)
	})

	t.Run("Existing chart - invalid", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *platformApi.ArangoPlatformService) {
			obj.Spec.Chart = &sharedApi.Object{
				Name: "secret",
			}
		})

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.False(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)

		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.Nil(t, release)
	})

	t.Run("Install Release", func(t *testing.T) {
		// Test
		tests.Update(t, handler.kubeClient, handler.client, &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
			obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
		})

		require.NoError(t, tests.Handle(chartHandler, tests.NewItem(t, operation.Update, chart)))
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.Equal(t, chart.Status.Info.Checksum, extension.Status.Conditions.Hash(platformApi.ChartFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)

		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.NotNil(t, release)

		require.Equal(t, release.Version, extension.Status.Release.Version)
		require.Equal(t, 1, extension.Status.Release.Version)

		cm := suite.GetConfigMap(t, handler.kubeClient, ns, "secret", "example")
		require.NotNil(t, cm)
		require.Equal(t, "PLACEHOLDER", cm.Data)
	})

	t.Run("Delete Release", func(t *testing.T) {
		// Test
		tests.DeleteObjects(t, handler.kubeClient, handler.client, &extension)

		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Validate
		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.Nil(t, release)

		cm := suite.GetConfigMap(t, handler.kubeClient, ns, "secret", "example")
		require.Nil(t, cm)
	})

	t.Run("Re-Install Release", func(t *testing.T) {
		// Arrange
		extension = tests.NewMetaObject[*platformApi.ArangoPlatformService](t, ns, "example",
			func(t *testing.T, obj *platformApi.ArangoPlatformService) {
				obj.Spec.Chart = &sharedApi.Object{
					Name: "secret",
				}
				obj.Spec.Deployment = &sharedApi.Object{
					Name: "example",
				}
			})

		tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.Equal(t, chart.Status.Info.Checksum, extension.Status.Conditions.Hash(platformApi.ChartFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)

		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.NotNil(t, release)

		require.Equal(t, release.Version, extension.Status.Release.Version)
		require.Equal(t, 1, extension.Status.Release.Version)

		cm := suite.GetConfigMap(t, handler.kubeClient, ns, "secret", "example")
		require.NotNil(t, cm)
		require.Equal(t, "PLACEHOLDER", cm.Data)
	})

	t.Run("Re-Install Release if Deleted manually", func(t *testing.T) {
		// Arrange
		_, err := handler.helm.Uninstall(context.Background(), "example")
		require.NoError(t, err)

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.Equal(t, chart.Status.Info.Checksum, extension.Status.Conditions.Hash(platformApi.ChartFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)

		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.NotNil(t, release)

		require.Equal(t, release.Version, extension.Status.Release.Version)
		require.Equal(t, 1, extension.Status.Release.Version)

		cm := suite.GetConfigMap(t, handler.kubeClient, ns, "secret", "example")
		require.NotNil(t, cm)
		require.Equal(t, "PLACEHOLDER", cm.Data)
	})

	t.Run("Upgrade Release if Values Changed", func(t *testing.T) {
		// Arrange
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *platformApi.ArangoPlatformService) {
			obj.Spec.Values = sharedApi.NewAnyT(t, map[string]string{
				"foo": "bar",
			})
		})

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.Equal(t, chart.Status.Info.Checksum, extension.Status.Conditions.Hash(platformApi.ChartFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)

		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.NotNil(t, release)

		require.Equal(t, release.Version, extension.Status.Release.Version)
		require.Equal(t, 2, extension.Status.Release.Version)

		cm := suite.GetConfigMap(t, handler.kubeClient, ns, "secret", "example")
		require.NotNil(t, cm)
		require.Equal(t, "PLACEHOLDER", cm.Data)
	})

	t.Run("Ensure Import Works", func(t *testing.T) {
		// Arrange
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *platformApi.ArangoPlatformService) {
			obj.Finalizers = nil
		})
		tests.DeleteObjects(t, handler.kubeClient, handler.client, &extension)

		extension = tests.NewMetaObject[*platformApi.ArangoPlatformService](t, ns, "example",
			func(t *testing.T, obj *platformApi.ArangoPlatformService) {
				obj.Spec.Chart = &sharedApi.Object{
					Name: "secret",
				}
				obj.Spec.Deployment = &sharedApi.Object{
					Name: "example",
				}
				obj.Spec.Values = sharedApi.NewAnyT(t, map[string]string{
					"foo": "bar",
				})
			})

		tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.Equal(t, chart.Status.Info.Checksum, extension.Status.Conditions.Hash(platformApi.ChartFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)

		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.NotNil(t, release)

		require.Equal(t, release.Version, extension.Status.Release.Version)
		require.Equal(t, 2, extension.Status.Release.Version)

		cm := suite.GetConfigMap(t, handler.kubeClient, ns, "secret", "example")
		require.NotNil(t, cm)
		require.Equal(t, "PLACEHOLDER", cm.Data)
	})

	t.Run("Overrides - Ensure Service Overrides are in place", func(t *testing.T) {
		// Arrange
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *platformApi.ArangoPlatformService) {
			obj.Spec.Values = sharedApi.NewAnyT(t, map[string]string{
				"data": "service",
			})
		})

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.Equal(t, chart.Status.Info.Checksum, extension.Status.Conditions.Hash(platformApi.ChartFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)

		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.NotNil(t, release)

		require.Equal(t, release.Version, extension.Status.Release.Version)
		require.Equal(t, 3, extension.Status.Release.Version)

		cm := suite.GetConfigMap(t, handler.kubeClient, ns, "secret", "example")
		require.NotNil(t, cm)
		require.Equal(t, "service", cm.Data)
	})

	t.Run("Overrides - Ensure Chart Overrides does not take presence over Service", func(t *testing.T) {
		// Arrange
		tests.Update(t, handler.kubeClient, handler.client, &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
			obj.Spec.Overrides = sharedApi.NewAnyT(t, map[string]string{
				"data": "chart",
			})
		})

		// Test
		require.NoError(t, tests.Handle(chartHandler, tests.NewItem(t, operation.Update, chart)))
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.Equal(t, chart.Status.Info.Checksum, extension.Status.Conditions.Hash(platformApi.ChartFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)

		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.NotNil(t, release)

		require.Equal(t, release.Version, extension.Status.Release.Version)
		require.Equal(t, 3, extension.Status.Release.Version)

		cm := suite.GetConfigMap(t, handler.kubeClient, ns, "secret", "example")
		require.NotNil(t, cm)
		require.Equal(t, "service", cm.Data)
	})

	t.Run("Overrides - Ensure Chart Overrides are used", func(t *testing.T) {
		// Arrange
		tests.Update(t, handler.kubeClient, handler.client, &extension, func(t *testing.T, obj *platformApi.ArangoPlatformService) {
			obj.Spec.Values = nil
		})

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Validate
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.DeploymentFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ChartFoundCondition))
		require.Equal(t, chart.Status.Info.Checksum, extension.Status.Conditions.Hash(platformApi.ChartFoundCondition))
		require.True(t, extension.Status.Conditions.IsTrue(platformApi.ReadyCondition))
		require.NotNil(t, extension.Status.Deployment)

		release, err := handler.helm.Status(context.Background(), "example")
		require.NoError(t, err)
		require.NotNil(t, release)

		require.Equal(t, release.Version, extension.Status.Release.Version)
		require.Equal(t, 4, extension.Status.Release.Version)

		cm := suite.GetConfigMap(t, handler.kubeClient, ns, "secret", "example")
		require.NotNil(t, cm)
		require.Equal(t, "chart", cm.Data)
	})
}
