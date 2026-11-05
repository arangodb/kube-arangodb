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
	"math"
	goStrings "strings"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1/definition"
	pbMetaV1 "github.com/arangodb/kube-arangodb/integrations/meta/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// jobStore persists jobs in the MetaStore via the MetaV1 gRPC client.
//
// Storage layout, modelled on the ArangoDB Agency supervision job lifecycle
// (jobs relocate between per-state collections):
//
//	links:<link_id>:jobs:<bucket>:<inv_priority>:<job_id>
//
// where <bucket> is the job's current state (todo/pending/finished/failed/cancelled).
// The <bucket> segment is the authoritative source of the job's current state; the
// job object's status history is kept as a best-effort audit trail. State transitions
// relocate the object between buckets with the MetaV1 Move operation (atomic, within a
// single transaction), so a pending job is claimed by exactly one handler.
//
// <inv_priority> is the priority inverted to a fixed width so that a prefix List of the
// todo bucket returns the highest-priority jobs first. Because the job id is prefixed
// with a fixed-width creation epoch, ties within a priority are broken FIFO (oldest first).
type jobStore struct {
	lock      sync.Mutex
	meta      pbMetaV1.MetaV1Client
	linkID    string
	handlerID string
}

func newJobStore(meta pbMetaV1.MetaV1Client, linkID, handlerID string) *jobStore {
	return &jobStore{
		meta:      meta,
		linkID:    linkID,
		handlerID: handlerID,
	}
}

const (
	bucketTodo      = "todo"
	bucketPending   = "pending"
	bucketFinished  = "finished"
	bucketFailed    = "failed"
	bucketCancelled = "cancelled"
)

// stateBucket maps a JobState to its storage bucket segment.
func stateBucket(s pbLinkV1.JobState) string {
	switch s {
	case pbLinkV1.JobState_JOB_STATE_TODO:
		return bucketTodo
	case pbLinkV1.JobState_JOB_STATE_PENDING:
		return bucketPending
	case pbLinkV1.JobState_JOB_STATE_FINISHED:
		return bucketFinished
	case pbLinkV1.JobState_JOB_STATE_FAILED:
		return bucketFailed
	case pbLinkV1.JobState_JOB_STATE_CANCELLED:
		return bucketCancelled
	default:
		return bucketTodo
	}
}

// bucketState is the reverse of stateBucket.
func bucketState(bucket string) pbLinkV1.JobState {
	switch bucket {
	case bucketPending:
		return pbLinkV1.JobState_JOB_STATE_PENDING
	case bucketFinished:
		return pbLinkV1.JobState_JOB_STATE_FINISHED
	case bucketFailed:
		return pbLinkV1.JobState_JOB_STATE_FAILED
	case bucketCancelled:
		return pbLinkV1.JobState_JOB_STATE_CANCELLED
	default:
		return pbLinkV1.JobState_JOB_STATE_TODO
	}
}

// invPriority inverts a priority into a fixed-width-friendly value so that ascending key
// ordering yields the highest priority first. Negative priorities are treated as 0.
func invPriority(priority int32) int64 {
	if priority < 0 {
		priority = 0
	}
	return int64(math.MaxInt32) - int64(priority)
}

func (s *jobStore) jobsPrefix() string {
	return fmt.Sprintf("links:%s:jobs:", s.linkID)
}

func (s *jobStore) bucketPrefix(bucket string) string {
	return fmt.Sprintf("links:%s:jobs:%s:", s.linkID, bucket)
}

// jobKey builds the storage key for a job in the given bucket.
func (s *jobStore) jobKey(bucket string, priority int32, jobID string) string {
	return fmt.Sprintf("%s%010d:%s", s.bucketPrefix(bucket), invPriority(priority), jobID)
}

func handlerKey(linkID, handlerID string) string {
	return fmt.Sprintf("links:%s:handlers:%s", linkID, handlerID)
}

// FileStorePath returns the FileStore path for a job's results
func FileStorePath(linkID, jobID string) string {
	return fmt.Sprintf("/links/%s/%s/", linkID, jobID)
}

// newJobID builds a job id prefixed with a fixed-width creation epoch (nanoseconds) followed
// by a uuid. The epoch prefix makes ids sort chronologically, which gives FIFO ordering
// within a priority; the uuid suffix guarantees uniqueness.
func newJobID(created *timestamppb.Timestamp) string {
	return fmt.Sprintf("%019d-%s", created.AsTime().UnixNano(), uuid.New().String())
}

func (s *jobStore) Create(ctx context.Context, job *pbLinkV1.Job) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.set(ctx, s.jobKey(bucketTodo, job.GetPriority(), job.GetId()), job, "")
}

func (s *jobStore) Get(ctx context.Context, id string) (*pbLinkV1.Job, string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	job, _, rev, err := s.locate(ctx, id)
	return job, rev, err
}

// set marshals and stores the job at the given key, optionally guarded by a revision.
func (s *jobStore) set(ctx context.Context, key string, job *pbLinkV1.Job, rev string) error {
	obj, err := anypb.New(job)
	if err != nil {
		return errors.Errorf("failed to marshal job: %v", err)
	}

	req := &pbMetaV1.SetRequest{
		Key:    key,
		Object: obj,
	}
	if rev != "" {
		req.Revision = &rev
	}

	if _, err := s.meta.Set(ctx, req); err != nil {
		return err
	}

	return nil
}

// getKey fetches the job stored at the exact key and reconciles its reported state with
// the authoritative bucket encoded in the key.
func (s *jobStore) getKey(ctx context.Context, key string) (*pbLinkV1.Job, string, error) {
	resp, err := s.meta.Get(ctx, &pbMetaV1.ObjectRequest{Key: key})
	if err != nil {
		return nil, "", err
	}

	job, err := unmarshalJob(resp)
	if err != nil {
		return nil, "", err
	}

	reconcileState(job, s.bucketOf(key))

	return job, resp.GetRevision(), nil
}

// locate finds the current key of a job by id and returns the job, its key and revision.
func (s *jobStore) locate(ctx context.Context, id string) (*pbLinkV1.Job, string, string, error) {
	keys, err := s.listPrefix(ctx, s.jobsPrefix())
	if err != nil {
		return nil, "", "", err
	}

	suffix := ":" + id
	for _, key := range keys {
		if !goStrings.HasSuffix(key, suffix) {
			continue
		}

		job, rev, err := s.getKey(ctx, key)
		if err != nil {
			return nil, "", "", err
		}

		return job, key, rev, nil
	}

	return nil, "", "", status.Errorf(codes.NotFound, "job %s not found", id)
}

// bucketOf extracts the bucket segment from a job key.
func (s *jobStore) bucketOf(key string) string {
	rest := goStrings.TrimPrefix(key, s.jobsPrefix())
	if i := goStrings.IndexByte(rest, ':'); i >= 0 {
		return rest[:i]
	}
	return rest
}

func (s *jobStore) listPrefix(ctx context.Context, prefix string) ([]string, error) {
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

// ListJobs returns all jobs, optionally filtered by state. When a state filter is set only
// that state's bucket is listed, otherwise every bucket is scanned.
func (s *jobStore) ListJobs(ctx context.Context, filter *pbLinkV1.JobState) ([]*pbLinkV1.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	prefix := s.jobsPrefix()
	if filter != nil {
		prefix = s.bucketPrefix(stateBucket(*filter))
	}

	keys, err := s.listPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	var jobs []*pbLinkV1.Job
	for _, key := range keys {
		job, _, err := s.getKey(ctx, key)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// PickUp atomically claims the highest-priority pending job and moves it to the Pending
// (in-progress) bucket. The Move operation relocates the job within a single transaction,
// so exactly one handler can claim a given job. Returns nil if no todo jobs are available.
func (s *jobStore) PickUp(ctx context.Context) (*pbLinkV1.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// The todo bucket is already ordered by (priority desc, creation asc).
	keys, err := s.listPrefix(ctx, s.bucketPrefix(bucketTodo))
	if err != nil {
		return nil, err
	}

	todoPrefix := s.bucketPrefix(bucketTodo)
	pendingPrefix := s.bucketPrefix(bucketPending)

	for _, key := range keys {
		dst := pendingPrefix + goStrings.TrimPrefix(key, todoPrefix)

		resp, err := s.meta.Move(ctx, &pbMetaV1.MoveRequest{
			Source:      key,
			Destination: dst,
		})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// Another handler already claimed this job; try the next one.
				continue
			}
			return nil, err
		}

		job, err := unmarshalJob(resp)
		if err != nil {
			return nil, err
		}

		// Annotate the claimed job. The job already lives in the pending bucket, which is
		// authoritative, so a failure here does not wedge it (it stays claimed).
		job.HandlerId = util.NewType(s.handlerID)
		pushStatus(job, &pbLinkV1.JobStatus{
			State:       pbLinkV1.JobState_JOB_STATE_PENDING,
			Description: "Job picked up",
		})
		job.Result = util.NewType(FileStorePath(s.linkID, job.GetId()))

		if err := s.set(ctx, dst, job, resp.GetRevision()); err != nil {
			return nil, err
		}

		return job, nil
	}

	return nil, nil
}

// UpdateStatus transitions a job to a new state, relocating it between buckets and
// validating the transition.
func (s *jobStore) UpdateStatus(ctx context.Context, id string, statusUpdate *pbLinkV1.JobStatus) (*pbLinkV1.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	job, key, rev, err := s.locate(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := validateTransition(currentState(job), statusUpdate.State); err != nil {
		return nil, err
	}

	return s.transition(ctx, job, key, rev, statusUpdate)
}

// Cancel moves a job to Cancelled if it is in the Todo or Pending state.
func (s *jobStore) Cancel(ctx context.Context, id string) (*pbLinkV1.Job, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	job, key, rev, err := s.locate(ctx, id)
	if err != nil {
		return nil, err
	}

	switch currentState(job) {
	case pbLinkV1.JobState_JOB_STATE_TODO,
		pbLinkV1.JobState_JOB_STATE_PENDING:
	default:
		return nil, errors.Errorf("cannot cancel job in state %s", currentState(job).String())
	}

	return s.transition(ctx, job, key, rev, &pbLinkV1.JobStatus{
		State:       pbLinkV1.JobState_JOB_STATE_CANCELLED,
		Description: "Cancelled by user",
	})
}

// transition pushes the new status onto the job and relocates it to the destination bucket.
// When the destination bucket differs from the current one the object is relocated with an
// atomic, revision-guarded Move; a status update that stays within the same bucket is a
// plain revision-guarded write.
func (s *jobStore) transition(ctx context.Context, job *pbLinkV1.Job, key, rev string, statusUpdate *pbLinkV1.JobStatus) (*pbLinkV1.Job, error) {
	pushStatus(job, statusUpdate)

	dst := s.jobKey(stateBucket(statusUpdate.State), job.GetPriority(), job.GetId())

	if dst == key {
		if err := s.set(ctx, key, job, rev); err != nil {
			return nil, err
		}
		return job, nil
	}

	resp, err := s.meta.Move(ctx, &pbMetaV1.MoveRequest{
		Source:      key,
		Destination: dst,
		Revision:    util.NewType(rev),
	})
	if err != nil {
		return nil, err
	}

	if err := s.set(ctx, dst, job, resp.GetRevision()); err != nil {
		return nil, err
	}

	return job, nil
}

func validateTransition(from, to pbLinkV1.JobState) error {
	switch from {
	case pbLinkV1.JobState_JOB_STATE_PENDING:
		switch to {
		case pbLinkV1.JobState_JOB_STATE_PENDING,
			pbLinkV1.JobState_JOB_STATE_FINISHED,
			pbLinkV1.JobState_JOB_STATE_FAILED,
			pbLinkV1.JobState_JOB_STATE_CANCELLED:
			return nil
		}
	case pbLinkV1.JobState_JOB_STATE_TODO:
		if to == pbLinkV1.JobState_JOB_STATE_CANCELLED {
			return nil
		}
	}
	return errors.Errorf("invalid state transition from %s to %s", from.String(), to.String())
}

const maxStatusHistory = 10

// currentState returns the current state of a job from its status history.
func currentState(job *pbLinkV1.Job) pbLinkV1.JobState {
	if len(job.Statuses) == 0 {
		return pbLinkV1.JobState_JOB_STATE_TODO
	}
	return job.Statuses[0].State
}

// reconcileState ensures the job's reported current state matches the authoritative bucket.
// This only diverges if a post-Move annotation write did not complete; in that case the
// bucket wins and a synthetic status entry is prepended.
func reconcileState(job *pbLinkV1.Job, bucket string) {
	want := bucketState(bucket)
	if currentState(job) == want {
		return
	}
	pushStatus(job, &pbLinkV1.JobStatus{
		State:       want,
		Description: "Recovered from store state",
	})
}

// pushStatus prepends a new status to the job's history, keeping at most maxStatusHistory entries.
func pushStatus(job *pbLinkV1.Job, s *pbLinkV1.JobStatus) {
	if s.Updated == nil {
		s.Updated = timestamppb.Now()
	}
	job.Statuses = append([]*pbLinkV1.JobStatus{s}, job.Statuses...)
	if len(job.Statuses) > maxStatusHistory {
		job.Statuses = job.Statuses[:maxStatusHistory]
	}
}

func unmarshalJob(resp *pbMetaV1.ObjectResponse) (*pbLinkV1.Job, error) {
	obj := resp.GetObject()
	if obj == nil {
		return nil, errors.Errorf("empty object in response")
	}

	var job pbLinkV1.Job
	if err := obj.UnmarshalTo(&job); err != nil {
		return nil, errors.Errorf("failed to unmarshal job: %v", err)
	}

	return &job, nil
}
