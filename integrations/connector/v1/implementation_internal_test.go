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

	pbConnectorV1 "github.com/arangodb/kube-arangodb/integrations/connector/v1/definition"
)

func Test_Internal_PickUpJob_Empty(t *testing.T) {
	impl := newTestImpl(t)

	resp, err := impl.PickUpJob(context.Background(), nil)
	require.NoError(t, err)
	require.Nil(t, resp.Id)
}

func Test_Internal_PickUpJob_Success(t *testing.T) {
	impl := newTestImpl(t)

	id := createTestJob(t, impl, "test")
	pickedID := pickUp(t, impl)
	require.Equal(t, id, pickedID)

	job := requireJobState(t, impl, id, pbConnectorV1.JobState_JOB_STATE_SCHEDULED)
	require.NotNil(t, job.HandlerId)
	require.NotEmpty(t, *job.HandlerId)
	require.NotNil(t, job.Result)
	require.Contains(t, *job.Result, testConnectorID)
	require.Contains(t, *job.Result, id)

	requireStatusHistory(t, job,
		pbConnectorV1.JobState_JOB_STATE_SCHEDULED,
		pbConnectorV1.JobState_JOB_STATE_PENDING,
	)
}

func Test_Internal_PickUpJob_SecondReturnsEmpty(t *testing.T) {
	impl := newTestImpl(t)

	createTestJob(t, impl, "only-one")
	pickUp(t, impl)

	resp, err := impl.PickUpJob(context.Background(), nil)
	require.NoError(t, err)
	require.Nil(t, resp.Id)
}

func Test_Internal_FullLifecycle(t *testing.T) {
	impl := newTestImpl(t)

	id := createTestJob(t, impl, `{"aql": "FOR d IN col RETURN d"}`)

	// Pending → Scheduled
	pickUp(t, impl)

	// Scheduled → Running
	updateStatus(t, impl, id, pbConnectorV1.JobState_JOB_STATE_RUNNING, "Executing query")

	// Running → Completed
	job := updateStatus(t, impl, id, pbConnectorV1.JobState_JOB_STATE_COMPLETED, "Done")

	requireStatusHistory(t, job,
		pbConnectorV1.JobState_JOB_STATE_COMPLETED,
		pbConnectorV1.JobState_JOB_STATE_RUNNING,
		pbConnectorV1.JobState_JOB_STATE_SCHEDULED,
		pbConnectorV1.JobState_JOB_STATE_PENDING,
	)
}

func Test_Internal_FailedJob(t *testing.T) {
	impl := newTestImpl(t)

	id := createTestJob(t, impl, "bad-query")
	pickUp(t, impl)
	updateStatus(t, impl, id, pbConnectorV1.JobState_JOB_STATE_RUNNING, "Executing")

	job := updateStatus(t, impl, id, pbConnectorV1.JobState_JOB_STATE_FAILED, "Syntax error")

	require.Equal(t, "Syntax error", job.Statuses[0].Description)
	requireStatusHistory(t, job,
		pbConnectorV1.JobState_JOB_STATE_FAILED,
		pbConnectorV1.JobState_JOB_STATE_RUNNING,
		pbConnectorV1.JobState_JOB_STATE_SCHEDULED,
		pbConnectorV1.JobState_JOB_STATE_PENDING,
	)
}

func Test_Internal_InvalidTransition(t *testing.T) {
	impl := newTestImpl(t)

	id := createTestJob(t, impl, "invalid")

	// Pending → Completed is invalid (must go through Scheduled)
	_, err := impl.UpdateJobStatus(context.Background(), &pbConnectorV1.UpdateJobStatusRequest{
		Id: id,
		Status: &pbConnectorV1.JobStatus{
			State: pbConnectorV1.JobState_JOB_STATE_COMPLETED,
		},
	})
	require.Error(t, err)
}

func Test_Internal_StatusHistoryMaxEntries(t *testing.T) {
	impl := newTestImpl(t)

	id := createTestJob(t, impl, "history")
	pickUp(t, impl)

	// Push many status updates to exceed maxStatusHistory
	for i := 0; i < 15; i++ {
		updateStatus(t, impl, id, pbConnectorV1.JobState_JOB_STATE_RUNNING, "Running")
		updateStatus(t, impl, id, pbConnectorV1.JobState_JOB_STATE_FAILED, "Failed")

		// Re-create and pick up fresh for next iteration since Failed is terminal
		id = createTestJob(t, impl, "history")
		pickUp(t, impl)
	}

	job := requireJobState(t, impl, id, pbConnectorV1.JobState_JOB_STATE_SCHEDULED)
	require.LessOrEqual(t, len(job.Statuses), maxStatusHistory)
}

func Test_Internal_UploadFile(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	id := createTestJob(t, impl, "upload")
	pickUp(t, impl)

	data := []byte(`{"count": 42}`)
	resp, err := impl.UploadFile(ctx, &pbConnectorV1.UploadFileRequest{
		JobId: id,
		Name:  "result.json",
		Data:  data,
	})
	require.NoError(t, err)
	require.Equal(t, int64(len(data)), resp.Bytes)
	require.NotEmpty(t, resp.Checksum)
}

func Test_Internal_UploadFile_MissingFields(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	_, err := impl.UploadFile(ctx, &pbConnectorV1.UploadFileRequest{})
	require.Error(t, err)
}

func Test_Internal_MultipleUploads(t *testing.T) {
	impl := newTestImpl(t)
	ctx := context.Background()

	id := createTestJob(t, impl, "multi-upload")
	pickUp(t, impl)

	// Upload multiple files
	files := map[string][]byte{
		"result.json":   []byte(`{"count": 42}`),
		"metadata.json": []byte(`{"format": "json"}`),
		"data.csv":      []byte("a,b,c\n1,2,3\n"),
	}

	for name, data := range files {
		resp, err := impl.UploadFile(ctx, &pbConnectorV1.UploadFileRequest{
			JobId: id,
			Name:  name,
			Data:  data,
		})
		require.NoError(t, err)
		require.Equal(t, int64(len(data)), resp.Bytes)
		require.NotEmpty(t, resp.Checksum)
	}
}
