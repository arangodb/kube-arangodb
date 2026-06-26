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

	pbLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1/definition"
)

func Test_ValidateTransition(t *testing.T) {
	tests := []struct {
		from  pbLinkV1.JobState
		to    pbLinkV1.JobState
		valid bool
	}{
		// From Scheduled
		{pbLinkV1.JobState_JOB_STATE_SCHEDULED, pbLinkV1.JobState_JOB_STATE_RUNNING, true},
		{pbLinkV1.JobState_JOB_STATE_SCHEDULED, pbLinkV1.JobState_JOB_STATE_FAILED, true},
		{pbLinkV1.JobState_JOB_STATE_SCHEDULED, pbLinkV1.JobState_JOB_STATE_CANCELLED, true},
		{pbLinkV1.JobState_JOB_STATE_SCHEDULED, pbLinkV1.JobState_JOB_STATE_COMPLETED, false},
		{pbLinkV1.JobState_JOB_STATE_SCHEDULED, pbLinkV1.JobState_JOB_STATE_PENDING, false},

		// From Running
		{pbLinkV1.JobState_JOB_STATE_RUNNING, pbLinkV1.JobState_JOB_STATE_COMPLETED, true},
		{pbLinkV1.JobState_JOB_STATE_RUNNING, pbLinkV1.JobState_JOB_STATE_FAILED, true},
		{pbLinkV1.JobState_JOB_STATE_RUNNING, pbLinkV1.JobState_JOB_STATE_CANCELLED, true},
		{pbLinkV1.JobState_JOB_STATE_RUNNING, pbLinkV1.JobState_JOB_STATE_PENDING, false},
		{pbLinkV1.JobState_JOB_STATE_RUNNING, pbLinkV1.JobState_JOB_STATE_SCHEDULED, false},

		// From Pending (only PickUp can move it)
		{pbLinkV1.JobState_JOB_STATE_PENDING, pbLinkV1.JobState_JOB_STATE_RUNNING, false},
		{pbLinkV1.JobState_JOB_STATE_PENDING, pbLinkV1.JobState_JOB_STATE_COMPLETED, false},

		// From terminal states
		{pbLinkV1.JobState_JOB_STATE_COMPLETED, pbLinkV1.JobState_JOB_STATE_RUNNING, false},
		{pbLinkV1.JobState_JOB_STATE_FAILED, pbLinkV1.JobState_JOB_STATE_RUNNING, false},
		{pbLinkV1.JobState_JOB_STATE_CANCELLED, pbLinkV1.JobState_JOB_STATE_RUNNING, false},
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
		job := &pbLinkV1.Job{}
		require.Equal(t, pbLinkV1.JobState_JOB_STATE_PENDING, currentState(job))
	})

	t.Run("with status", func(t *testing.T) {
		job := &pbLinkV1.Job{
			Statuses: []*pbLinkV1.JobStatus{
				{State: pbLinkV1.JobState_JOB_STATE_RUNNING},
				{State: pbLinkV1.JobState_JOB_STATE_SCHEDULED},
				{State: pbLinkV1.JobState_JOB_STATE_PENDING},
			},
		}
		require.Equal(t, pbLinkV1.JobState_JOB_STATE_RUNNING, currentState(job))
	})
}

func Test_PushStatus(t *testing.T) {
	t.Run("appends to front", func(t *testing.T) {
		job := &pbLinkV1.Job{
			Statuses: []*pbLinkV1.JobStatus{
				{State: pbLinkV1.JobState_JOB_STATE_PENDING, Description: "created"},
			},
		}
		pushStatus(job, &pbLinkV1.JobStatus{
			State:       pbLinkV1.JobState_JOB_STATE_SCHEDULED,
			Description: "scheduled",
		})

		require.Len(t, job.Statuses, 2)
		require.Equal(t, pbLinkV1.JobState_JOB_STATE_SCHEDULED, job.Statuses[0].State)
		require.Equal(t, pbLinkV1.JobState_JOB_STATE_PENDING, job.Statuses[1].State)
		require.NotNil(t, job.Statuses[0].Updated)
	})

	t.Run("caps at max", func(t *testing.T) {
		job := &pbLinkV1.Job{}
		for i := 0; i < 15; i++ {
			pushStatus(job, &pbLinkV1.JobStatus{
				State:       pbLinkV1.JobState_JOB_STATE_RUNNING,
				Description: "iteration",
			})
		}
		require.Len(t, job.Statuses, maxStatusHistory)
	})
}

func Test_FileStorePath(t *testing.T) {
	path := FileStorePath("connector-123", "job-456")
	require.Equal(t, "/links/connector-123/job-456/", path)
}
