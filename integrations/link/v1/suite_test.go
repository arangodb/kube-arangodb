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
	"bytes"
	"context"
	"encoding/json"
	goStrings "strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1/definition"
	pbStorageV2 "github.com/arangodb/kube-arangodb/integrations/storage/v2/definition"
	testIntegration "github.com/arangodb/kube-arangodb/pkg/util/tests/integration"
)

const testLinkID = "00000000-0000-0000-0000-000000000001"

type testEnv struct {
	*implementation
	storage pbStorageV2.StorageV2Client
	t       *testing.T
}

func newTestEnv(t *testing.T) *testEnv {
	storage := testIntegration.NewStorageV2Client(t)
	impl := New(testIntegration.NewMetaV1Client(), storage, testLinkID).(*implementation)
	return &testEnv{implementation: impl, storage: storage, t: t}
}

// ListFiles returns all object paths stored under the job's prefix.
func (e *testEnv) ListFiles(jobID string) []string {
	e.t.Helper()
	return storageList(e.t, context.Background(), e.storage, storagePrefix(testLinkID, jobID))
}

// ReadFile returns the raw bytes of a file stored for the given job.
func (e *testEnv) ReadFile(jobID, name string) []byte {
	e.t.Helper()
	return storageRead(e.t, context.Background(), e.storage, storagePrefix(testLinkID, jobID)+name)
}

// ReadFileJSON reads a file from storage and unmarshals it into T.
func ReadFileJSON[T any](e *testEnv, jobID, name string) T {
	e.t.Helper()
	return storageReadJSON[T](e.t, context.Background(), e.storage, storagePrefix(testLinkID, jobID)+name)
}

func newTestImpl(t *testing.T) *implementation {
	return newTestEnv(t).implementation
}

// storagePrefix normalises a FileStorePath for use with the test storage mock.
// The mock's ListObjects returns paths relative to its root (no leading '/'),
// so we strip the leading '/' from the prefix produced by FileStorePath.
func storagePrefix(linkID, jobID string) string { //nolint:unparam
	return goStrings.TrimPrefix(FileStorePath(linkID, jobID), "/")
}

// storageList returns all object paths under the given prefix via pbStorageV2.List.
func storageList(t *testing.T, ctx context.Context, storage pbStorageV2.StorageV2Client, prefix string) []string {
	t.Helper()

	objects, err := pbStorageV2.List(ctx, storage, prefix)
	require.NoError(t, err)

	var names []string
	for _, obj := range objects {
		names = append(names, obj.GetPath().GetPath())
	}
	return names
}

// storageRead reads an object from storage and returns its raw bytes.
func storageRead(t *testing.T, ctx context.Context, storage pbStorageV2.StorageV2Client, key string) []byte {
	t.Helper()

	var buf bytes.Buffer
	_, err := pbStorageV2.Receive(ctx, storage, key, &buf)
	require.NoError(t, err)
	return buf.Bytes()
}

// storageReadJSON reads an object from storage, unmarshals it into out, and returns it.
func storageReadJSON[T any](t *testing.T, ctx context.Context, storage pbStorageV2.StorageV2Client, key string) T {
	t.Helper()

	data := storageRead(t, ctx, storage, key)

	var out T
	require.NoError(t, json.Unmarshal(data, &out))
	return out
}

func createTestJob(t *testing.T, impl *implementation, query string) string {
	t.Helper()

	ctx := context.Background()
	resp, err := impl.CreateJob(ctx, &pbLinkV1.CreateJobRequest{
		Input: []byte(query),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Id)
	return resp.Id
}

func requireJobState(t *testing.T, impl *implementation, id string, state pbLinkV1.JobState) *pbLinkV1.Job {
	t.Helper()

	job, err := impl.GetJob(context.Background(), &pbLinkV1.GetJobRequest{Id: id})
	require.NoError(t, err)
	require.Equal(t, state, currentState(job))
	return job
}

func requireStatusHistory(t *testing.T, job *pbLinkV1.Job, states ...pbLinkV1.JobState) {
	t.Helper()

	require.Len(t, job.Statuses, len(states))
	for i, s := range states {
		require.Equal(t, s, job.Statuses[i].State, "status[%d]", i)
		require.NotNil(t, job.Statuses[i].Updated, "status[%d].updated_at", i)
	}
}

func pickUp(t *testing.T, impl *implementation) string {
	t.Helper()

	resp, err := impl.PickUpJob(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, resp.Id)
	return *resp.Id
}

func updateStatus(t *testing.T, impl *implementation, id string, state pbLinkV1.JobState, desc string) *pbLinkV1.Job {
	t.Helper()

	resp, err := impl.UpdateJobStatus(context.Background(), &pbLinkV1.UpdateJobStatusRequest{
		Id: id,
		Status: &pbLinkV1.JobStatus{
			State:       state,
			Description: desc,
			Updated:     timestamppb.Now(),
		},
	})
	require.NoError(t, err)
	return resp.Job
}
