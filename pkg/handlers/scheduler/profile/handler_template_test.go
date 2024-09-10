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

package profile

import (
	"testing"

	"github.com/stretchr/testify/require"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod"
	schedulerPodResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod/resources"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Handler_Template(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoProfile](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec.Template = &schedulerApi.ProfileTemplate{}
		})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.True(t, extension.Status.Conditions.IsTrue(schedulerApi.SpecValidCondition))
	require.True(t, extension.Status.Conditions.IsTrue(schedulerApi.ReadyCondition))

	require.NotNil(t, extension.Status.Accepted)
	require.Equal(t, extension.Spec.Template, extension.Status.Accepted.Template)
	require.Equal(t, "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a", extension.Status.Accepted.Checksum)
}

func Test_Handler_InvalidTemplate(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoProfile](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec.Template = &schedulerApi.ProfileTemplate{
				Container: &schedulerApi.ProfileContainerTemplate{
					Containers: schedulerContainerApi.Containers{
						"^%$@#": {},
					},
				},
			}
		})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	// Test
	require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

	// Refresh
	refresh(t)

	// Assert
	require.False(t, extension.Status.Conditions.IsTrue(schedulerApi.SpecValidCondition))
	require.False(t, extension.Status.Conditions.IsTrue(schedulerApi.ReadyCondition))

	require.Nil(t, extension.Status.Accepted)
}

func Test_Handler_Template_Update(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoProfile](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec.Template = &schedulerApi.ProfileTemplate{}
		})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	t.Run("First", func(t *testing.T) {
		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Assert
		require.True(t, extension.Status.Conditions.IsTrue(schedulerApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(schedulerApi.ReadyCondition))

		require.NotNil(t, extension.Status.Accepted)
		require.Equal(t, extension.Spec.Template, extension.Status.Accepted.Template)
		require.Equal(t, "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a", extension.Status.Accepted.Checksum)

		require.Nil(t, extension.Status.Accepted.Template.Pod)
	})

	t.Run("Second", func(t *testing.T) {
		// Arrange
		extension.Spec.Template = &schedulerApi.ProfileTemplate{
			Pod: &schedulerPodApi.Pod{
				Image: &schedulerPodResourcesApi.Image{
					ImagePullSecrets: schedulerPodResourcesApi.ImagePullSecrets{
						"test",
					},
				},
			},
		}

		tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Assert
		require.True(t, extension.Status.Conditions.IsTrue(schedulerApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(schedulerApi.ReadyCondition))

		require.NotNil(t, extension.Status.Accepted)
		require.Equal(t, extension.Spec.Template, extension.Status.Accepted.Template)
		require.Equal(t, "f48e2e30be82828a43873befdf9eb877ee4d34d45b2eb45a2f253020955e022a", extension.Status.Accepted.Checksum)

		require.NotNil(t, extension.Status.Accepted.Template.Pod)
		require.Equal(t, &schedulerPodApi.Pod{
			Image: &schedulerPodResourcesApi.Image{
				ImagePullSecrets: schedulerPodResourcesApi.ImagePullSecrets{
					"test",
				},
			},
		}, extension.Status.Accepted.Template.Pod)
	})
}

func Test_Handler_Template_UpdateInvalid(t *testing.T) {
	handler := newFakeHandler()

	// Arrange
	extension := tests.NewMetaObject[*schedulerApi.ArangoProfile](t, tests.FakeNamespace, "test",
		func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec.Template = &schedulerApi.ProfileTemplate{}
		})

	refresh := tests.CreateObjects(t, handler.kubeClient, handler.client, &extension)

	t.Run("First", func(t *testing.T) {
		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Assert
		require.True(t, extension.Status.Conditions.IsTrue(schedulerApi.SpecValidCondition))
		require.True(t, extension.Status.Conditions.IsTrue(schedulerApi.ReadyCondition))

		require.NotNil(t, extension.Status.Accepted)
		require.Equal(t, extension.Spec.Template, extension.Status.Accepted.Template)
		require.Equal(t, "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a", extension.Status.Accepted.Checksum)

		require.Nil(t, extension.Status.Accepted.Template.Pod)
	})

	t.Run("Second", func(t *testing.T) {
		// Arrange
		extension.Spec.Template = &schedulerApi.ProfileTemplate{
			Container: &schedulerApi.ProfileContainerTemplate{
				Containers: schedulerContainerApi.Containers{
					"^%$@#": {},
				},
			},
		}

		tests.UpdateObjects(t, handler.kubeClient, handler.client, &extension)

		// Test
		require.NoError(t, tests.Handle(handler, tests.NewItem(t, operation.Update, extension)))

		// Refresh
		refresh(t)

		// Assert
		require.False(t, extension.Status.Conditions.IsTrue(schedulerApi.SpecValidCondition))
		require.False(t, extension.Status.Conditions.IsTrue(schedulerApi.ReadyCondition))

		require.Nil(t, extension.Status.Accepted)
	})
}
