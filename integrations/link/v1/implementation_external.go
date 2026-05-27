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

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbLinkV1 "github.com/arangodb/kube-arangodb/integrations/link/v1/definition"
)

func (i *implementation) CreateJob(ctx context.Context, req *pbLinkV1.CreateJobRequest) (*pbLinkV1.CreateJobResponse, error) {
	id := uuid.New().String()

	job := &pbLinkV1.Job{
		Id:      id,
		LinkId:  i.linkID,
		Query:   req.GetQuery(),
		Timeout: req.Timeout,
		Statuses: []*pbLinkV1.JobStatus{
			{
				State:       pbLinkV1.JobState_JOB_STATE_PENDING,
				Description: "Job created",
				Updated:     timestamppb.Now(),
			},
		},
		Created: timestamppb.Now(),
	}

	if err := i.store.Create(ctx, job); err != nil {
		return nil, err
	}

	return &pbLinkV1.CreateJobResponse{Id: id}, nil
}

func (i *implementation) ListJobs(ctx context.Context, req *pbLinkV1.ListJobsRequest) (*pbLinkV1.ListJobsResponse, error) {
	keys, err := i.store.List(ctx)
	if err != nil {
		return nil, err
	}

	prefix := i.store.jobKeyPrefix()
	var jobs []*pbLinkV1.Job
	for _, key := range keys {
		id := key[len(prefix):]
		job, _, err := i.store.Get(ctx, id)
		if err != nil {
			continue
		}

		if req.State != nil && currentState(job) != *req.State {
			continue
		}

		jobs = append(jobs, job)
	}

	return &pbLinkV1.ListJobsResponse{Jobs: jobs}, nil
}

func (i *implementation) CancelJob(ctx context.Context, req *pbLinkV1.CancelJobRequest) (*pbLinkV1.CancelJobResponse, error) {
	job, err := i.store.Cancel(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &pbLinkV1.CancelJobResponse{Job: job}, nil
}

func (i *implementation) GetInfo(ctx context.Context, req *pbLinkV1.GetInfoRequest) (*pbLinkV1.GetInfoResponse, error) {
	return &pbLinkV1.GetInfoResponse{Info: i.info}, nil
}
