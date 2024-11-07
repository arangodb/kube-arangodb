//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package deployment

import (
	"testing"

	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod"
	schedulerPodResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod/resources"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Handler_Create(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerDeployment](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerDeployment) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "df484e6467f4b2445aae3a93ed7da1e374a689ac7e616c86ba31a3b2dc3e3244", extension.Status.Object.GetChecksum())
}

func Test_Handler_Update(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerDeployment](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerDeployment) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "df484e6467f4b2445aae3a93ed7da1e374a689ac7e616c86ba31a3b2dc3e3244", extension.Status.Object.GetChecksum())

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerDeployment) {
		obj.Spec.Replicas = util.NewType[int32](2)
	})
	tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "6ce82430599f0e4a3dc3983226076179a27559f0f1d87194eaa3c2d482aaceb3", extension.Status.Object.GetChecksum())
}

func Test_Handler_Recreate(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerDeployment](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerDeployment) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "df484e6467f4b2445aae3a93ed7da1e374a689ac7e616c86ba31a3b2dc3e3244", extension.Status.Object.GetChecksum())

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerDeployment) {
		obj.Status.Object.UID = util.NewType[types.UID]("TEST")
	})
	tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "df484e6467f4b2445aae3a93ed7da1e374a689ac7e616c86ba31a3b2dc3e3244", extension.Status.Object.GetChecksum())
}

func Test_Handler_Parent(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerDeployment](t, tests.FakeNamespace, "test", func(t *testing.T, obj *schedulerApi.ArangoSchedulerDeployment) {})
	deployment := tests.NewMetaObject[*apps.Deployment](t, tests.FakeNamespace, "test", func(t *testing.T, obj *apps.Deployment) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &deployment)

	require.Len(t, deployment.OwnerReferences, 1)
}

func Test_Handler_Propagate(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerDeployment](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerDeployment) {})
	deployment := tests.NewMetaObject[*apps.Deployment](t, tests.FakeNamespace, "test", func(t *testing.T, obj *apps.Deployment) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &deployment)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "df484e6467f4b2445aae3a93ed7da1e374a689ac7e616c86ba31a3b2dc3e3244", extension.Status.Object.GetChecksum())
	require.EqualValues(t, 0, extension.Status.Replicas)

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerDeployment) {
		obj.Spec.Replicas = util.NewType[int32](2)
	})
	tests.Apply(t, deployment, func(t *testing.T, obj *apps.Deployment) {
		obj.Status.Replicas = 5
	})
	tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension, &deployment)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &deployment)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "6ce82430599f0e4a3dc3983226076179a27559f0f1d87194eaa3c2d482aaceb3", extension.Status.Object.GetChecksum())
	require.EqualValues(t, 5, extension.Status.Replicas)
}

func Test_Handler_Profile(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	profile := tests.NewMetaObject[*schedulerApi.ArangoProfile](t, tests.FakeNamespace, "test", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
		obj.Spec.Template = &schedulerApi.ProfileTemplate{
			Pod: &schedulerPodApi.Pod{
				Volumes: &schedulerPodResourcesApi.Volumes{
					Volumes: []core.Volume{
						{
							Name:         "test",
							VolumeSource: core.VolumeSource{},
						},
					},
				},
			},
		}
	}, tests.MarkArangoProfileAsReady)
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerDeployment](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerDeployment) {
			obj.Spec.Profiles = []string{profile.GetName()}
			obj.Spec.DeploymentSpec.Replicas = util.NewType[int32](10)
		})
	deployment := tests.NewMetaObject[*apps.Deployment](t, tests.FakeNamespace, "test")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.EqualError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)), "Profile with name `test` is missing")

	tests.CreateObjects(t, handler.kubeClient, handler.client, &profile)
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &deployment)
	require.NotNil(t, deployment)

	require.NotNil(t, deployment.Spec.Replicas)
	require.EqualValues(t, 10, *deployment.Spec.Replicas)

	require.Len(t, extension.Status.Profiles, 1)
	require.Equal(t, profile.GetName(), extension.Status.Profiles[0])
	require.Len(t, deployment.Spec.Template.Spec.Volumes, 1)
	require.EqualValues(t, "test", deployment.Spec.Template.Spec.Volumes[0].Name)
}
