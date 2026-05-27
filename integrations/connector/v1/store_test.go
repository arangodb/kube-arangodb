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
	"testing"

	"github.com/stretchr/testify/require"

	pbConnectorV1 "github.com/arangodb/kube-arangodb/integrations/connector/v1/definition"
)

func Test_ValidateTransition(t *testing.T) {
	tests := []struct {
		from  pbConnectorV1.JobState
		to    pbConnectorV1.JobState
		valid bool
	}{
		// From Scheduled
		{pbConnectorV1.JobState_JOB_STATE_SCHEDULED, pbConnectorV1.JobState_JOB_STATE_RUNNING, true},
		{pbConnectorV1.JobState_JOB_STATE_SCHEDULED, pbConnectorV1.JobState_JOB_STATE_FAILED, true},
		{pbConnectorV1.JobState_JOB_STATE_SCHEDULED, pbConnectorV1.JobState_JOB_STATE_CANCELLED, true},
		{pbConnectorV1.JobState_JOB_STATE_SCHEDULED, pbConnectorV1.JobState_JOB_STATE_COMPLETED, false},
		{pbConnectorV1.JobState_JOB_STATE_SCHEDULED, pbConnectorV1.JobState_JOB_STATE_PENDING, false},

		// From Running
		{pbConnectorV1.JobState_JOB_STATE_RUNNING, pbConnectorV1.JobState_JOB_STATE_COMPLETED, true},
		{pbConnectorV1.JobState_JOB_STATE_RUNNING, pbConnectorV1.JobState_JOB_STATE_FAILED, true},
		{pbConnectorV1.JobState_JOB_STATE_RUNNING, pbConnectorV1.JobState_JOB_STATE_CANCELLED, true},
		{pbConnectorV1.JobState_JOB_STATE_RUNNING, pbConnectorV1.JobState_JOB_STATE_PENDING, false},
		{pbConnectorV1.JobState_JOB_STATE_RUNNING, pbConnectorV1.JobState_JOB_STATE_SCHEDULED, false},

		// From Pending (only PickUp can move it)
		{pbConnectorV1.JobState_JOB_STATE_PENDING, pbConnectorV1.JobState_JOB_STATE_RUNNING, false},
		{pbConnectorV1.JobState_JOB_STATE_PENDING, pbConnectorV1.JobState_JOB_STATE_COMPLETED, false},

		// From terminal states
		{pbConnectorV1.JobState_JOB_STATE_COMPLETED, pbConnectorV1.JobState_JOB_STATE_RUNNING, false},
		{pbConnectorV1.JobState_JOB_STATE_FAILED, pbConnectorV1.JobState_JOB_STATE_RUNNING, false},
		{pbConnectorV1.JobState_JOB_STATE_CANCELLED, pbConnectorV1.JobState_JOB_STATE_RUNNING, false},
	}

	for _, tt := range tests {
		t.Run(tt.from.String()+"_to_"+tt.to.String(), func(t *testing.T) {
			err := validateTransition(tt.from, tt.to)
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func Test_CurrentState(t *testing.T) {
	t.Run("empty status", func(t *testing.T) {
		job := &pbConnectorV1.Job{}
		require.Equal(t, pbConnectorV1.JobState_JOB_STATE_PENDING, currentState(job))
	})

	t.Run("with status", func(t *testing.T) {
		job := &pbConnectorV1.Job{
			Statuses: []*pbConnectorV1.JobStatus{
				{State: pbConnectorV1.JobState_JOB_STATE_RUNNING},
				{State: pbConnectorV1.JobState_JOB_STATE_SCHEDULED},
				{State: pbConnectorV1.JobState_JOB_STATE_PENDING},
			},
		}
		require.Equal(t, pbConnectorV1.JobState_JOB_STATE_RUNNING, currentState(job))
	})
}

func Test_PushStatus(t *testing.T) {
	t.Run("appends to front", func(t *testing.T) {
		job := &pbConnectorV1.Job{
			Statuses: []*pbConnectorV1.JobStatus{
				{State: pbConnectorV1.JobState_JOB_STATE_PENDING, Description: "created"},
			},
		}
		pushStatus(job, &pbConnectorV1.JobStatus{
			State:       pbConnectorV1.JobState_JOB_STATE_SCHEDULED,
			Description: "scheduled",
		})

		require.Len(t, job.Statuses, 2)
		require.Equal(t, pbConnectorV1.JobState_JOB_STATE_SCHEDULED, job.Statuses[0].State)
		require.Equal(t, pbConnectorV1.JobState_JOB_STATE_PENDING, job.Statuses[1].State)
		require.NotNil(t, job.Statuses[0].Updated)
	})

	t.Run("caps at max", func(t *testing.T) {
		job := &pbConnectorV1.Job{}
		for i := 0; i < 15; i++ {
			pushStatus(job, &pbConnectorV1.JobStatus{
				State:       pbConnectorV1.JobState_JOB_STATE_RUNNING,
				Description: "iteration",
			})
		}
		require.Len(t, job.Statuses, maxStatusHistory)
	})
}

func Test_FileStorePath(t *testing.T) {
	path := FileStorePath("connector-123", "job-456")
	require.Equal(t, "/connectors/connector-123/job-456/", path)
}
