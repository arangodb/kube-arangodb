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

package cronjob

import (
	"testing"

	"github.com/stretchr/testify/require"
	batch "k8s.io/api/batch/v1"
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
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerCronJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerCronJob) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "2f9e4c718f8bf1f0880e64aa44c10142acb59ca88a4c08d89ab7daadc93b115e", extension.Status.Object.GetChecksum())
}

func Test_Handler_Update(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerCronJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerCronJob) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "2f9e4c718f8bf1f0880e64aa44c10142acb59ca88a4c08d89ab7daadc93b115e", extension.Status.Object.GetChecksum())

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerCronJob) {
		obj.Spec.StartingDeadlineSeconds = util.NewType[int64](2)
	})
	tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "5411ac008bce56d38b0d0e36d8bbbbb904c02c01dc3e8052f4467d6f24f9c7b5", extension.Status.Object.GetChecksum())
}

func Test_Handler_Recreate(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerCronJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerCronJob) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "2f9e4c718f8bf1f0880e64aa44c10142acb59ca88a4c08d89ab7daadc93b115e", extension.Status.Object.GetChecksum())

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerCronJob) {
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
	require.Equal(t, "2f9e4c718f8bf1f0880e64aa44c10142acb59ca88a4c08d89ab7daadc93b115e", extension.Status.Object.GetChecksum())
}

func Test_Handler_Parent(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerCronJob](t, tests.FakeNamespace, "test", func(t *testing.T, obj *schedulerApi.ArangoSchedulerCronJob) {})
	cronJob := tests.NewMetaObject[*batch.CronJob](t, tests.FakeNamespace, "test")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &cronJob)

	require.Len(t, cronJob.OwnerReferences, 1)
}

func Test_Handler_Propagate(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerCronJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerCronJob) {})
	cronJob := tests.NewMetaObject[*batch.CronJob](t, tests.FakeNamespace, "test")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &cronJob)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "2f9e4c718f8bf1f0880e64aa44c10142acb59ca88a4c08d89ab7daadc93b115e", extension.Status.Object.GetChecksum())
	require.Nil(t, cronJob.Spec.StartingDeadlineSeconds)
	require.Len(t, extension.Status.Active, 0)

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerCronJob) {
		obj.Spec.StartingDeadlineSeconds = util.NewType[int64](2)
	})
	tests.Apply(t, cronJob, func(t *testing.T, obj *batch.CronJob) {
		obj.Status.Active = []core.ObjectReference{{}}
	})
	tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension, &cronJob)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &cronJob)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "5411ac008bce56d38b0d0e36d8bbbbb904c02c01dc3e8052f4467d6f24f9c7b5", extension.Status.Object.GetChecksum())
	require.NotNil(t, cronJob.Spec.StartingDeadlineSeconds)
	require.EqualValues(t, 2, *cronJob.Spec.StartingDeadlineSeconds)
	require.Len(t, extension.Status.Active, 1)
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
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerCronJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerCronJob) {
			obj.Spec.Profiles = []string{profile.GetName()}
		})
	cronJob := tests.NewMetaObject[*batch.CronJob](t, tests.FakeNamespace, "test")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.EqualError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)), "Profile with name `test` is missing")

	tests.CreateObjects(t, handler.kubeClient, handler.client, &profile)
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &cronJob)
	require.NotNil(t, cronJob)

	require.Len(t, extension.Status.Profiles, 1)
	require.Equal(t, profile.GetName(), extension.Status.Profiles[0])

	require.Len(t, extension.Status.Profiles, 1)
	require.Equal(t, profile.GetName(), extension.Status.Profiles[0])
	require.Len(t, cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes, 1)
	require.EqualValues(t, "test", cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes[0].Name)
}
