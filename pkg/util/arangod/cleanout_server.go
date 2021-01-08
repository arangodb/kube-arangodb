//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package arangod

import (
	"context"
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
)

// CleanoutJobStatus is a strongly typed status of an agency cleanout-server-job.
type CleanoutJobStatus struct {
	state  string
	reason string
}

// IsFailed returns true when the job is failed
func (s CleanoutJobStatus) IsFailed() bool {
	return s.state == "Failed"
}

// IsFinished returns true when the job is finished
func (s CleanoutJobStatus) IsFinished() bool {
	return s.state == "Finished"
}

// Reason returns the reason for the current state.
func (s CleanoutJobStatus) Reason() string {
	return s.reason
}

// String returns a string representation of the given state.
func (s CleanoutJobStatus) String() string {
	return fmt.Sprintf("state: '%s', reason: '%s'", s.state, s.reason)
}

var (
	agencyJobStateKeyPrefixes = [][]string{
		{"arango", "Target", "ToDo"},
		{"arango", "Target", "Pending"},
		{"arango", "Target", "Finished"},
		{"arango", "Target", "Failed"},
	}
)

type agencyJob struct {
	Reason string `json:"reason,omitempty"`
	Server string `json:"server,omitempty"`
	JobID  string `json:"jobId,omitempty"`
	Type   string `json:"type,omitempty"`
}

const (
	agencyJobTypeCleanOutServer = "cleanOutServer"
)

// CleanoutServerJobStatus checks the status of a cleanout-server job with given ID.
func CleanoutServerJobStatus(ctx context.Context, jobID string, client driver.Client, agencyClient agency.Agency) (CleanoutJobStatus, error) {
	for _, keyPrefix := range agencyJobStateKeyPrefixes {
		key := append(keyPrefix, jobID)
		var job agencyJob
		if err := agencyClient.ReadKey(ctx, key, &job); err == nil {
			return CleanoutJobStatus{
				state:  keyPrefix[len(keyPrefix)-1],
				reason: job.Reason,
			}, nil
		} else if agency.IsKeyNotFound(err) {
			continue
		} else {
			return CleanoutJobStatus{}, errors.WithStack(err)
		}
	}
	// Job not found in any states
	return CleanoutJobStatus{
		reason: "job not found",
	}, nil
}
