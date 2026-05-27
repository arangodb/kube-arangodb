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
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbConnectorV1 "github.com/arangodb/kube-arangodb/integrations/connector/v1/definition"
	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// jobStore persists jobs in MetaStore via the MetaV1 gRPC client.
// Keys: connectors/<connector_id>/<job_id>
// Uses a local lock to serialize operations within a single instance,
// and revision-based optimistic concurrency for safety across instances.
type jobStore struct {
	lock        sync.Mutex
	meta        pbMetaV1.MetaV1Client
	connectorID string
	handlerID   string
}

func newJobStore(meta pbMetaV1.MetaV1Client, connectorID, handlerID string) *jobStore {
	return &jobStore{
		meta:        meta,
		connectorID: connectorID,
		handlerID:   handlerID,
	}
}

func (s *jobStore) jobKey(jobID string) string {
	return fmt.Sprintf("connectors/%s/jobs/%s", s.connectorID, jobID)
}

func (s *jobStore) jobKeyPrefix() string {
	return fmt.Sprintf("connectors/%s/jobs/", s.connectorID)
}

func handlerKey(connectorID, handlerID string) string {
	return fmt.Sprintf("connectors/%s/handlers/%s", connectorID, handlerID)
}

// FileStorePath returns the FileStore path for a job's results
func FileStorePath(connectorID, jobID string) string {
	return fmt.Sprintf("/connectors/%s/%s/", connectorID, jobID)
}

func (s *jobStore) Create(ctx context.Context, job *pbConnectorV1.Job) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	obj, err := anypb.New(job)
	if err != nil {
		return errors.Errorf("failed to marshal job: %v", err)
	}

	_, err = s.meta.Set(ctx, &pbMetaV1.SetRequest{
		Key:    s.jobKey(job.Id),
		Object: obj,
	})
	return err
}

func (s *jobStore) Get(ctx context.Context, id string) (*pbConnectorV1.Job, string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.get(ctx, id)
}

func (s *jobStore) get(ctx context.Context, id string) (*pbConnectorV1.Job, string, error) {
	resp, err := s.meta.Get(ctx, &pbMetaV1.ObjectRequest{
		Key: s.jobKey(id),
	})
	if err != nil {
		return nil, "", err
	}

	job, err := unmarshalJob(resp)
	if err != nil {
		return nil, "", err
	}

	return job, resp.GetRevision(), nil
}

func (s *jobStore) List(ctx context.Context) ([]string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.list(ctx)
}

func (s *jobStore) list(ctx context.Context) ([]string, error) {
	prefix := s.jobKeyPrefix()
	stream, err := s.meta.List(ctx, &pbMetaV1.ListRequest{
		Prefix: util.NewType(prefix),
	})
	if err != nil {
		return nil, err
	}

	var keys []string
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		keys = append(keys, chunk.Keys...)
	}
	return keys, nil
}

func (s *jobStore) update(ctx context.Context, job *pbConnectorV1.Job, rev string) error {
	obj, err := anypb.New(job)
	if err != nil {
		return errors.Errorf("failed to marshal job: %v", err)
	}

	_, err = s.meta.Set(ctx, &pbMetaV1.SetRequest{
		Key:      s.jobKey(job.Id),
		Revision: &rev,
		Object:   obj,
	})
	return err
}

// PickUp atomically finds one Pending job and moves it to Scheduled.
// Sets the handler_id to this instance's HandlerUUID.
// Uses revision check to ensure only one connector instance picks up the job.
// Returns nil if no pending jobs available.
func (s *jobStore) PickUp(ctx context.Context) (*pbConnectorV1.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	keys, err := s.list(ctx)
	if err != nil {
		return nil, err
	}

	prefix := s.jobKeyPrefix()

	for _, key := range keys {
		// Strip prefix to get the job ID
		id := key[len(prefix):]

		job, rev, err := s.get(ctx, id)
		if err != nil {
			continue
		}

		if currentState(job) != pbConnectorV1.JobState_JOB_STATE_PENDING {
			continue
		}

		// Attempt atomic transition Pending → Scheduled using revision
		job.HandlerId = util.NewType(s.handlerID)
		pushStatus(job, &pbConnectorV1.JobStatus{
			State:       pbConnectorV1.JobState_JOB_STATE_SCHEDULED,
			Description: "Job scheduled",
		})
		job.Result = util.NewType(FileStorePath(s.connectorID, job.Id))

		if err := s.update(ctx, job, rev); err != nil {
			// Revision conflict — another instance picked it up, try next
			continue
		}

		return job, nil
	}

	return nil, nil
}

// UpdateStatus updates the status of a job with revision check.
// Validates state transitions before updating.
func (s *jobStore) UpdateStatus(ctx context.Context, id string, status *pbConnectorV1.JobStatus) (*pbConnectorV1.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	job, rev, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(currentState(job), status.State); err != nil {
		return nil, err
	}

	pushStatus(job, status)

	if err := s.update(ctx, job, rev); err != nil {
		return nil, err
	}

	return job, nil
}

// Cancel moves a job to Cancelled if it is in Pending, Scheduled, or Running state.
// Uses revision check for concurrency safety.
func (s *jobStore) Cancel(ctx context.Context, id string) (*pbConnectorV1.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	job, rev, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}

	switch currentState(job) {
	case pbConnectorV1.JobState_JOB_STATE_PENDING,
		pbConnectorV1.JobState_JOB_STATE_SCHEDULED,
		pbConnectorV1.JobState_JOB_STATE_RUNNING:
	default:
		return nil, errors.Errorf("cannot cancel job in state %s", currentState(job).String())
	}

	pushStatus(job, &pbConnectorV1.JobStatus{
		State:       pbConnectorV1.JobState_JOB_STATE_CANCELLED,
		Description: "Cancelled by user",
	})

	if err := s.update(ctx, job, rev); err != nil {
		return nil, err
	}

	return job, nil
}

func validateTransition(from, to pbConnectorV1.JobState) error {
	switch from {
	case pbConnectorV1.JobState_JOB_STATE_SCHEDULED:
		if to == pbConnectorV1.JobState_JOB_STATE_RUNNING ||
			to == pbConnectorV1.JobState_JOB_STATE_FAILED ||
			to == pbConnectorV1.JobState_JOB_STATE_CANCELLED {
			return nil
		}
	case pbConnectorV1.JobState_JOB_STATE_RUNNING:
		if to == pbConnectorV1.JobState_JOB_STATE_COMPLETED ||
			to == pbConnectorV1.JobState_JOB_STATE_FAILED ||
			to == pbConnectorV1.JobState_JOB_STATE_CANCELLED {
			return nil
		}
	}
	return errors.Errorf("invalid state transition from %s to %s", from.String(), to.String())
}

const maxStatusHistory = 10

// currentState returns the current state of a job from its status history.
func currentState(job *pbConnectorV1.Job) pbConnectorV1.JobState {
	if len(job.Statuses) == 0 {
		return pbConnectorV1.JobState_JOB_STATE_PENDING
	}
	return job.Statuses[0].State
}

// pushStatus prepends a new status to the job's history, keeping at most maxStatusHistory entries.
func pushStatus(job *pbConnectorV1.Job, s *pbConnectorV1.JobStatus) {
	if s.Updated == nil {
		s.Updated = timestamppb.Now()
	}
	job.Statuses = append([]*pbConnectorV1.JobStatus{s}, job.Statuses...)
	if len(job.Statuses) > maxStatusHistory {
		job.Statuses = job.Statuses[:maxStatusHistory]
	}
}

func unmarshalJob(resp *pbMetaV1.ObjectResponse) (*pbConnectorV1.Job, error) {
	obj := resp.GetObject()
	if obj == nil {
		return nil, errors.Errorf("empty object in response")
	}

	var job pbConnectorV1.Job
	if err := obj.UnmarshalTo(&job); err != nil {
		return nil, errors.Errorf("failed to unmarshal job: %v", err)
	}

	return &job, nil
}
