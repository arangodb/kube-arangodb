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

package batchjob

import (
	"testing"

	"github.com/stretchr/testify/require"
	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Handler_Create(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerBatchJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerBatchJob) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "c12919994bb3b13dfc1cd7903bd2020a4da93064d93b068171d1567a203c62c4", extension.Status.Object.GetChecksum())
}

func Test_Handler_Update(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerBatchJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerBatchJob) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "c12919994bb3b13dfc1cd7903bd2020a4da93064d93b068171d1567a203c62c4", extension.Status.Object.GetChecksum())

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerBatchJob) {
		obj.Spec.Completions = util.NewType[int32](2)
	})
	tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "65257e9b53283da2bfc00caeca08eee9cfbc465a3032119cb95c113efdf62b25", extension.Status.Object.GetChecksum())
}

func Test_Handler_Recreate(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerBatchJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerBatchJob) {})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "c12919994bb3b13dfc1cd7903bd2020a4da93064d93b068171d1567a203c62c4", extension.Status.Object.GetChecksum())

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerBatchJob) {
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
	require.Equal(t, "c12919994bb3b13dfc1cd7903bd2020a4da93064d93b068171d1567a203c62c4", extension.Status.Object.GetChecksum())
}

func Test_Handler_Parent(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerBatchJob](t, tests.FakeNamespace, "test", func(t *testing.T, obj *schedulerApi.ArangoSchedulerBatchJob) {})
	batchJob := tests.NewMetaObject[*batch.Job](t, tests.FakeNamespace, "test")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Validate
	require.NotNil(t, extension.Status.Object)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &batchJob)

	require.Len(t, batchJob.OwnerReferences, 1)
}

func Test_Handler_Propagate(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerBatchJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerBatchJob) {})
	batchJob := tests.NewMetaObject[*batch.Job](t, tests.FakeNamespace, "test")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &batchJob)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "c12919994bb3b13dfc1cd7903bd2020a4da93064d93b068171d1567a203c62c4", extension.Status.Object.GetChecksum())
	require.Nil(t, batchJob.Spec.Completions)
	require.EqualValues(t, 0, extension.Status.Active)

	// Update
	tests.Apply(t, extension, func(t *testing.T, obj *schedulerApi.ArangoSchedulerBatchJob) {
		obj.Spec.Completions = util.NewType[int32](2)
	})
	tests.Apply(t, batchJob, func(t *testing.T, obj *batch.Job) {
		obj.Status.Active = 1
	})
	tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension, &batchJob)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &batchJob)

	// Validate
	require.NotNil(t, extension.Status.Object)
	require.Equal(t, extension.GetName(), extension.Status.Object.GetName())
	require.Equal(t, "65257e9b53283da2bfc00caeca08eee9cfbc465a3032119cb95c113efdf62b25", extension.Status.Object.GetChecksum())
	require.NotNil(t, batchJob.Spec.Completions)
	require.EqualValues(t, 2, *batchJob.Spec.Completions)
	require.EqualValues(t, 1, extension.Status.Active)
}

func Test_Handler_Profile(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	profile := tests.NewMetaObject[*schedulerApi.ArangoProfile](t, tests.FakeNamespace, "test", tests.MarkArangoProfileAsReady)
	extension := tests.NewMetaObject[*schedulerApi.ArangoSchedulerBatchJob](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoSchedulerBatchJob) {
			obj.Spec.Profiles = []string{profile.GetName()}
		})
	batchJob := tests.NewMetaObject[*batch.Job](t, tests.FakeNamespace, "test")

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.EqualError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)), "Profile with name `test` is missing")

	tests.CreateObjects(t, handler.kubeClient, handler.client, &profile)
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)
	tests.RefreshObjects(t, handler.kubeClient, handler.client, &batchJob)
	require.NotNil(t, batchJob)

	require.Len(t, extension.Status.Profiles, 1)
	require.Equal(t, profile.GetName(), extension.Status.Profiles[0])
}
