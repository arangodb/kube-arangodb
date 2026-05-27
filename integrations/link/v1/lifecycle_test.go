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

// Test_Lifecycle_Complete tests the full happy path:
// CreateJob → PickUp → Running → Upload files → Completed
func Test_Lifecycle_Complete(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	// 1. Create job
	id := createTestJob(t, impl, `{"aql": "FOR d IN users RETURN d"}`)
	requireJobState(t, impl, id, pbLinkV1.JobState_JOB_STATE_PENDING)

	// 2. Pick up
	pickedID := pickUp(t, impl)
	require.Equal(t, id, pickedID)
	job := requireJobState(t, impl, id, pbLinkV1.JobState_JOB_STATE_SCHEDULED)
	require.NotNil(t, job.HandlerId)
	require.NotNil(t, job.Result)

	// 3. Running
	updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_RUNNING, "Executing AQL query")
	requireJobState(t, impl, id, pbLinkV1.JobState_JOB_STATE_RUNNING)

	// 4. Upload result files
	resultData := []byte(`[{"name":"alice"},{"name":"bob"}]`)
	resp, err := impl.UploadFile(ctx, &pbLinkV1.UploadFileRequest{
		JobId: id,
		Name:  "result.json",
		Data:  resultData,
	})
	require.NoError(t, err)
	require.Equal(t, int64(len(resultData)), resp.Bytes)
	require.NotEmpty(t, resp.Checksum)

	resp, err = impl.UploadFile(ctx, &pbLinkV1.UploadFileRequest{
		JobId: id,
		Name:  "stats.json",
		Data:  []byte(`{"count":2,"time_ms":15}`),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Checksum)

	// 5. Complete
	job = updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_COMPLETED, "Query returned 2 documents")

	requireStatusHistory(t, job,
		pbLinkV1.JobState_JOB_STATE_COMPLETED,
		pbLinkV1.JobState_JOB_STATE_RUNNING,
		pbLinkV1.JobState_JOB_STATE_SCHEDULED,
		pbLinkV1.JobState_JOB_STATE_PENDING,
	)
}

// Test_Lifecycle_Failed tests the failure path:
// CreateJob → PickUp → Running → Failed
func Test_Lifecycle_Failed(t *testing.T) {
	impl := newTestImpl(t)

	id := createTestJob(t, impl, `{"aql": "INVALID SYNTAX"}`)
	pickUp(t, impl)
	updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_RUNNING, "Executing")

	job := updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_FAILED, "AQL parse error at position 1:8")

	require.Equal(t, "AQL parse error at position 1:8", job.Statuses[0].Description)
	requireStatusHistory(t, job,
		pbLinkV1.JobState_JOB_STATE_FAILED,
		pbLinkV1.JobState_JOB_STATE_RUNNING,
		pbLinkV1.JobState_JOB_STATE_SCHEDULED,
		pbLinkV1.JobState_JOB_STATE_PENDING,
	)
}

// Test_Lifecycle_CancelWhileRunning tests cancellation during execution:
// CreateJob → PickUp → Running → Cancel
func Test_Lifecycle_CancelWhileRunning(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	id := createTestJob(t, impl, `{"aql": "FOR d IN huge_collection RETURN d"}`)
	pickUp(t, impl)
	updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_RUNNING, "Executing long query")

	resp, err := impl.CancelJob(ctx, &pbLinkV1.CancelJobRequest{Id: id})
	require.NoError(t, err)

	requireStatusHistory(t, resp.Job,
		pbLinkV1.JobState_JOB_STATE_CANCELLED,
		pbLinkV1.JobState_JOB_STATE_RUNNING,
		pbLinkV1.JobState_JOB_STATE_SCHEDULED,
		pbLinkV1.JobState_JOB_STATE_PENDING,
	)
}

// Test_Lifecycle_MultipleJobs tests processing multiple jobs sequentially
func Test_Lifecycle_MultipleJobs(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	createTestJob(t, impl, "query1")
	createTestJob(t, impl, "query2")
	createTestJob(t, impl, "query3")

	// Pick up and complete all three (order is non-deterministic)
	for i := 0; i < 3; i++ {
		id := pickUp(t, impl)

		if i < 2 {
			updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_RUNNING, "Running")
			updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_COMPLETED, "Done")
		} else {
			updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_RUNNING, "Running")
			updateStatus(t, impl, id, pbLinkV1.JobState_JOB_STATE_FAILED, "Error")
		}
	}

	// List all — should be 3 jobs
	resp, err := impl.ListJobs(ctx, &pbLinkV1.ListJobsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Jobs, 3)

	// No more pending
	resp2, err := impl.PickUpJob(ctx, nil)
	require.NoError(t, err)
	require.Nil(t, resp2.Id)

	// Count by state
	completed := pbLinkV1.JobState_JOB_STATE_COMPLETED
	respCompleted, err := impl.ListJobs(ctx, &pbLinkV1.ListJobsRequest{State: &completed})
	require.NoError(t, err)
	require.Len(t, respCompleted.Jobs, 2)

	failed := pbLinkV1.JobState_JOB_STATE_FAILED
	respFailed, err := impl.ListJobs(ctx, &pbLinkV1.ListJobsRequest{State: &failed})
	require.NoError(t, err)
	require.Len(t, respFailed.Jobs, 1)
}
