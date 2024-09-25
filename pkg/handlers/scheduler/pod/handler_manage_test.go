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

package pod

import (
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Handler_Create(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerPod](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerPod) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "64759f2e87813091ac4dbb627ee7411316259132ca5a9603786993f122899c2c", extension.Status.Object.GetChecksum())
}

func Test_Handler_Update(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerPod](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerPod) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "64759f2e87813091ac4dbb627ee7411316259132ca5a9603786993f122899c2c", extension.Status.Object.GetChecksum())

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerPod) {
		obj.Spec.HostNetwork = true
	})
	tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "c982d1c7855103125df8330401d993eb1e8de85b2bd605ac61af3c872f4fa51d", extension.Status.Object.GetChecksum())
}

func Test_Handler_Recreate(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerPod](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerPod) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "64759f2e87813091ac4dbb627ee7411316259132ca5a9603786993f122899c2c", extension.Status.Object.GetChecksum())

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerPod) {
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
	require.Equal(t, "64759f2e87813091ac4dbb627ee7411316259132ca5a9603786993f122899c2c", extension.Status.Object.GetChecksum())
}

func Test_Handler_Parent(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerPod](t, tests.FakeNamespace, "test", func(t *testing.T, obj *schedulerApi.ArangoSchedulerPod) {})
	pod := tests.NewMetaObject[*core.Pod](t, tests.FakeNamespace, "test", func(t *testing.T, obj *core.Pod) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &pod)

	require.Len(t, pod.OwnerReferences, 1)
}

func Test_Handler_Propagate(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerPod](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerPod) {})
	pod := tests.NewMetaObject[*core.Pod](t, tests.FakeNamespace, "test", func(t *testing.T, obj *core.Pod) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &pod)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "64759f2e87813091ac4dbb627ee7411316259132ca5a9603786993f122899c2c", extension.Status.Object.GetChecksum())
	require.Equal(t, "", pod.Status.Message)

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerPod) {
		obj.Spec.HostNetwork = true
	})
	tests.Apply(t, pod, func(t *testing.T, obj *core.Pod) {
		obj.Status.Message = "RANDOM"
	})
	tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension, &pod)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &pod)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "c982d1c7855103125df8330401d993eb1e8de85b2bd605ac61af3c872f4fa51d", extension.Status.Object.GetChecksum())
	require.Equal(t, "RANDOM", pod.Status.Message)
}

func Test_Handler_Profile(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	profile := tests.NewMetaObject[*schedulerApi.ArangoProfile](t, tests.FakeNamespace, "test", tests.MarkArangoProfileAsReady)
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerPod](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerPod) {
			obj.Spec.Profiles = []string{profile.GetName()}
		})
	pod := tests.NewMetaObject[*core.Pod](t, tests.FakeNamespace, "test")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.EqualError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)), "Profile with name `test` is missing")

	tests.CreateObjects(t, handler.kubeClient, handler.client, &profile)
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &pod)
	require.NotNil(t, pod)

	require.Len(t, extension.Status.Profiles, 1)
	require.Equal(t, profile.GetName(), extension.Status.Profiles[0])
}
