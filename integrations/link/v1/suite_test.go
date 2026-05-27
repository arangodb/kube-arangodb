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
	"google.golang.org/protobuf/types/known/timestamppb"

	pbLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1/definition"
	testIntegration "github.com/arangodb/kube-arangodb/pkg/util/tests/integration"
)

const testLinkID = "00000000-0000-0000-0000-000000000001"

func newTestImpl(t *testing.T) *implementation {
	return New(testIntegration.NewMetaV1Client(), testIntegration.NewStorageV2Client(t), testLinkID).(*implementation)
}

func createTestJob(t *testing.T, impl *implementation, query string) string {
	t.Helper()

	ctx := context.Background()
	resp, err := impl.CreateJob(ctx, &pbLinkV1.CreateJobRequest{
		Query: []byte(query),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Id)
	return resp.Id
}

func requireJobState(t *testing.T, impl *implementation, id string, state pbLinkV1.JobState) *pbLinkV1.Job {
	t.Helper()

	resp, err := impl.GetJob(context.Background(), &pbLinkV1.GetJobRequest{Id: id})
	require.NoError(t, err)
	require.Equal(t, state, currentState(resp.Job))
	return resp.Job
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
