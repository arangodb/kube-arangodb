//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

	pbLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1/definition"
)

func Test_External_CreateJob(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	resp, err := impl.CreateJob(ctx, &pbLinkV1.CreateJobRequest{
		Input: []byte(`{"aql": "RETURN 1"}`),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Id)

	job := requireJobState(t, impl, resp.Id, pbLinkV1.JobState_JOB_STATE_TODO)
	require.Equal(t, testLinkID, job.LinkId)
	require.Equal(t, `{"aql": "RETURN 1"}`, string(job.Input))
	require.NotNil(t, job.Created)
	requireStatusHistory(t, job, pbLinkV1.JobState_JOB_STATE_TODO)
}

func Test_External_ListJobs(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	id1 := createTestJob(t, impl, "query1")
	id2 := createTestJob(t, impl, "query2")

	// List all
	resp, err := impl.ListJobs(ctx, &pbLinkV1.ListJobsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Jobs, 2)

	ids := []string{resp.Jobs[0].Id, resp.Jobs[1].Id}
	require.Contains(t, ids, id1)
	require.Contains(t, ids, id2)
}

func Test_External_ListJobs_FilterByState(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	createTestJob(t, impl, "todo")
	createTestJob(t, impl, "also-todo")

	// Pick up one → Pending
	pickedID := pickUp(t, impl)

	// List only Todo — should have 1 (the one not picked up)
	todo := pbLinkV1.JobState_JOB_STATE_TODO
	resp, err := impl.ListJobs(ctx, &pbLinkV1.ListJobsRequest{State: &todo})
	require.NoError(t, err)
	require.Len(t, resp.Jobs, 1)
	require.NotEqual(t, pickedID, resp.Jobs[0].Id)

	// List only Pending — should have 1 (the picked-up one)
	pending := pbLinkV1.JobState_JOB_STATE_PENDING
	resp, err = impl.ListJobs(ctx, &pbLinkV1.ListJobsRequest{State: &pending})
	require.NoError(t, err)
	require.Len(t, resp.Jobs, 1)
	require.Equal(t, pickedID, resp.Jobs[0].Id)
}

func Test_External_CancelJob_Pending(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	id := createTestJob(t, impl, "cancel-me")

	resp, err := impl.CancelJob(ctx, &pbLinkV1.CancelJobRequest{Id: id})
	require.NoError(t, err)
	require.Equal(t, pbLinkV1.JobState_JOB_STATE_CANCELLED, currentState(resp.Job))
}

func Test_External_CancelJob_Running(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	id := createTestJob(t, impl, "cancel-running")
	pickUp(t, impl)
	updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_PENDING, "Executing")

	resp, err := impl.CancelJob(ctx, &pbLinkV1.CancelJobRequest{Id: id})
	require.NoError(t, err)
	require.Equal(t, pbLinkV1.JobState_JOB_STATE_CANCELLED, currentState(resp.Job))
}

func Test_External_CancelJob_Completed_Fails(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	id := createTestJob(t, impl, "done")
	pickUp(t, impl)
	updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_PENDING, "Running")
	updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_FINISHED, "Done")

	_, err := impl.CancelJob(ctx, &pbLinkV1.CancelJobRequest{Id: id})
	require.Error(t, err)
}

func Test_External_GetJob_NotFound(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	_, err := impl.GetJob(ctx, &pbLinkV1.GetJobRequest{Id: "nonexistent"})
	require.Error(t, err)
}
