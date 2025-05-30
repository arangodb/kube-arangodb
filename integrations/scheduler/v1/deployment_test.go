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

package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Deployment(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := kclient.NewFakeClientBuilder().Add(
		tests.NewMetaObject(t, tests.FakeNamespace, "test", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec = schedulerApi.ProfileSpec{}
		}, tests.MarkArangoProfileAsReady),
		tests.NewMetaObject(t, tests.FakeNamespace, "test-select-all", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec = schedulerApi.ProfileSpec{
				Selectors: &schedulerApi.ProfileSelectors{
					Label: &meta.LabelSelector{},
				},
				Template: &schedulerApi.ProfileTemplate{},
			}
		}, tests.MarkArangoProfileAsReady),
		tests.NewMetaObject(t, tests.FakeNamespace, "test-select-specific", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec = schedulerApi.ProfileSpec{
				Selectors: &schedulerApi.ProfileSelectors{
					Label: &meta.LabelSelector{
						MatchLabels: map[string]string{
							"A": "B",
						},
					},
				},
				Template: &schedulerApi.ProfileTemplate{},
			}
		}, tests.MarkArangoProfileAsReady),
	).Client()

	scheduler := Client(t, ctx, client, func(c Configuration) Configuration {
		c.Namespace = tests.FakeNamespace
		c.VerifyAccess = false
		return c
	})

	t.Run("Ensure job does not exist - get", func(t *testing.T) {
		resp, err := scheduler.GetDeployment(context.Background(), &pbSchedulerV1.GetDeploymentRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.False(t, resp.GetExists())
	})

	t.Run("Ensure job does not exist - list", func(t *testing.T) {
		resp, err := scheduler.ListDeployment(context.Background(), &pbSchedulerV1.ListDeploymentRequest{})
		require.NoError(t, err)

		require.Len(t, resp.GetDeployments(), 0)
	})

	t.Run("Schedule Job", func(t *testing.T) {
		resp, err := scheduler.CreateDeployment(context.Background(), &pbSchedulerV1.CreateDeploymentRequest{
			Spec: &pbSchedulerV1.Spec{
				Metadata: &pbSchedulerV1.Metadata{
					Name: "test",
				},
				Base: &pbSchedulerV1.ObjectBase{
					Labels: nil,
					Profiles: []string{
						"test",
					},
				},
				Containers: map[string]*pbSchedulerV1.ContainerBase{
					"example": {
						Image: util.NewType("ubuntu:20.04"),
						Args: []string{
							"/bin/bash",
							"-c",
							"true",
						},
					},
				},
			},
			Deployment: &pbSchedulerV1.DeploymentSpec{
				Replicas: util.NewType[int32](5),
			},
		})
		require.NoError(t, err)

		require.EqualValues(t, "test", resp.GetName())
	})

	t.Run("Ensure job exist - get", func(t *testing.T) {
		resp, err := scheduler.GetDeployment(context.Background(), &pbSchedulerV1.GetDeploymentRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.True(t, resp.GetExists())
	})

	t.Run("Ensure job exist - list", func(t *testing.T) {
		resp, err := scheduler.ListDeployment(context.Background(), &pbSchedulerV1.ListDeploymentRequest{})
		require.NoError(t, err)

		require.Len(t, resp.GetDeployments(), 1)
		require.Contains(t, resp.GetDeployments(), "test")
	})

	t.Run("Ensure job details - pre", func(t *testing.T) {
		resp, err := scheduler.GetDeployment(context.Background(), &pbSchedulerV1.GetDeploymentRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.True(t, resp.GetExists())
		require.EqualValues(t, 0, resp.GetDeployment().GetStatus().GetReplicas())
	})

	t.Run("Ensure job details - update", func(t *testing.T) {
		job := tests.NewMetaObject[*schedulerApi.ArangoSchedulerDeployment](t, tests.FakeNamespace, "test")

		tests.RefreshObjectsC(t, client, &job)

		job.Status.Replicas = 1

		tests.UpdateObjectsC(t, client, &job)
	})

	t.Run("Ensure job details - post", func(t *testing.T) {
		resp, err := scheduler.GetDeployment(context.Background(), &pbSchedulerV1.GetDeploymentRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.True(t, resp.GetExists())
		require.EqualValues(t, 1, resp.GetDeployment().GetStatus().GetReplicas())
	})

	t.Run("Delete Job", func(t *testing.T) {
		resp, err := scheduler.DeleteDeployment(context.Background(), &pbSchedulerV1.DeleteDeploymentRequest{
			Name: "test",
		})
		require.NoError(t, err)
		require.True(t, resp.GetExists())
	})

	t.Run("Re-Delete Job", func(t *testing.T) {
		resp, err := scheduler.DeleteDeployment(context.Background(), &pbSchedulerV1.DeleteDeploymentRequest{
			Name: "test",
		})
		require.NoError(t, err)
		require.False(t, resp.GetExists())
	})

	t.Run("Ensure job does not exist after deletion - get", func(t *testing.T) {
		resp, err := scheduler.GetDeployment(context.Background(), &pbSchedulerV1.GetDeploymentRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.False(t, resp.GetExists())
	})

	t.Run("Ensure job does not exist after deletion - list", func(t *testing.T) {
		resp, err := scheduler.ListDeployment(context.Background(), &pbSchedulerV1.ListDeploymentRequest{})
		require.NoError(t, err)

		require.Len(t, resp.GetDeployments(), 0)
	})
}
