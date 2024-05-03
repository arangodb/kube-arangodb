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
	batch "k8s.io/api/batch/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_BatchJob(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := kclient.NewFakeClientBuilder().Add(
		tests.NewMetaObject(t, tests.FakeNamespace, "test", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec = schedulerApi.ProfileSpec{}
		}),
		tests.NewMetaObject(t, tests.FakeNamespace, "test-select-all", func(t *testing.T, obj *schedulerApi.ArangoProfile) {
			obj.Spec = schedulerApi.ProfileSpec{
				Selectors: &schedulerApi.ProfileSelectors{
					Label: &meta.LabelSelector{},
				},
				Template: &schedulerApi.ProfileTemplate{},
			}
		}),
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
		}),
	).Client()

	scheduler := Client(t, ctx, client, func(c Configuration) Configuration {
		c.Namespace = tests.FakeNamespace
		c.VerifyAccess = false
		return c
	})

	t.Run("Ensure job does not exist - get", func(t *testing.T) {
		resp, err := scheduler.GetBatchJob(context.Background(), &pbSchedulerV1.GetBatchJobRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.False(t, resp.GetExists())
	})

	t.Run("Ensure job does not exist - list", func(t *testing.T) {
		resp, err := scheduler.ListBatchJob(context.Background(), &pbSchedulerV1.ListBatchJobRequest{})
		require.NoError(t, err)

		require.Len(t, resp.GetBatchJobs(), 0)
	})

	t.Run("Schedule Job", func(t *testing.T) {
		resp, err := scheduler.CreateBatchJob(context.Background(), &pbSchedulerV1.CreateBatchJobRequest{
			Spec: &pbSchedulerV1.Spec{
				Metadata: &pbSchedulerV1.Metadata{
					Name: "test",
				},
				Job: &pbSchedulerV1.JobBase{
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
			BatchJob: &pbSchedulerV1.BatchJobSpec{
				Completions: util.NewType[int32](1),
			},
		})
		require.NoError(t, err)

		require.EqualValues(t, "test", resp.GetName())
		require.Len(t, resp.Profiles, 2)
		require.Contains(t, resp.Profiles, "test")
		require.Contains(t, resp.Profiles, "test-select-all")
		require.NotContains(t, resp.Profiles, "test-select-specific")
	})

	t.Run("Ensure job exist - get", func(t *testing.T) {
		resp, err := scheduler.GetBatchJob(context.Background(), &pbSchedulerV1.GetBatchJobRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.True(t, resp.GetExists())
	})

	t.Run("Ensure job exist - list", func(t *testing.T) {
		resp, err := scheduler.ListBatchJob(context.Background(), &pbSchedulerV1.ListBatchJobRequest{})
		require.NoError(t, err)

		require.Len(t, resp.GetBatchJobs(), 1)
		require.Contains(t, resp.GetBatchJobs(), "test")
	})

	t.Run("Ensure job details - pre", func(t *testing.T) {
		resp, err := scheduler.GetBatchJob(context.Background(), &pbSchedulerV1.GetBatchJobRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.True(t, resp.GetExists())
		require.EqualValues(t, 0, resp.GetBatchJob().GetStatus().GetSucceeded())
	})

	t.Run("Ensure job details - update", func(t *testing.T) {
		job := tests.NewMetaObject[*batch.Job](t, tests.FakeNamespace, "test")

		tests.RefreshObjectsC(t, client, &job)

		job.Status.Succeeded = 1

		tests.UpdateObjectsC(t, client, &job)
	})

	t.Run("Ensure job details - post", func(t *testing.T) {
		resp, err := scheduler.GetBatchJob(context.Background(), &pbSchedulerV1.GetBatchJobRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.True(t, resp.GetExists())
		require.EqualValues(t, 1, resp.GetBatchJob().GetStatus().GetSucceeded())
	})

	t.Run("Delete Job", func(t *testing.T) {
		resp, err := scheduler.DeleteBatchJob(context.Background(), &pbSchedulerV1.DeleteBatchJobRequest{
			Name: "test",
		})
		require.NoError(t, err)
		require.True(t, resp.GetExists())
	})

	t.Run("Re-Delete Job", func(t *testing.T) {
		resp, err := scheduler.DeleteBatchJob(context.Background(), &pbSchedulerV1.DeleteBatchJobRequest{
			Name: "test",
		})
		require.NoError(t, err)
		require.False(t, resp.GetExists())
	})

	t.Run("Ensure job does not exist after deletion - get", func(t *testing.T) {
		resp, err := scheduler.GetBatchJob(context.Background(), &pbSchedulerV1.GetBatchJobRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.False(t, resp.GetExists())
	})

	t.Run("Ensure job does not exist after deletion - list", func(t *testing.T) {
		resp, err := scheduler.ListBatchJob(context.Background(), &pbSchedulerV1.ListBatchJobRequest{})
		require.NoError(t, err)

		require.Len(t, resp.GetBatchJobs(), 0)
	})
}
