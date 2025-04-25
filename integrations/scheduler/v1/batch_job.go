//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/scheduler"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
)

func (i *implementation) CreateBatchJob(ctx context.Context, request *pbSchedulerV1.CreateBatchJobRequest) (*pbSchedulerV1.CreateBatchJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	template := scheduler.SpecAsTemplate(request.GetSpec())

	var spec schedulerApi.ArangoSchedulerBatchJob

	spec.Namespace = i.cfg.Namespace

	if meta := request.GetSpec().GetMetadata(); meta != nil {
		if util.TypeOrDefault(meta.GenerateName, false) {
			spec.GenerateName = meta.Name
		} else {
			spec.Name = meta.Name
		}
	}

	spec.Spec.Template = *template

	if batchJob := request.GetBatchJob(); batchJob != nil {
		if v := batchJob.Completions; v != nil {
			spec.Spec.Completions = v
		}

		if v := batchJob.Parallelism; v != nil {
			spec.Spec.Parallelism = v
		}

		if v := batchJob.BackoffLimit; v != nil {
			spec.Spec.BackoffLimit = v
		}
	}

	if batchJobSpec := request.GetSpec(); batchJobSpec != nil {
		if base := batchJobSpec.Base; base != nil {
			spec.Labels = base.Labels
		}
	}

	job, err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerBatchJobs(i.cfg.Namespace).Create(ctx, &spec, meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	return &pbSchedulerV1.CreateBatchJobResponse{
		Name: job.Name,
	}, nil
}

func (i *implementation) GetBatchJob(ctx context.Context, request *pbSchedulerV1.GetBatchJobRequest) (*pbSchedulerV1.GetBatchJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	job, err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerBatchJobs(i.cfg.Namespace).Get(ctx, request.GetName(), meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.GetBatchJobResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	return &pbSchedulerV1.GetBatchJobResponse{
		Exists: true,

		BatchJob: &pbSchedulerV1.BatchJob{
			Metadata: ExtractStatusMetadata(job.Status.ArangoSchedulerStatusMetadata),
			Spec: &pbSchedulerV1.BatchJobSpec{
				Parallelism:  job.Spec.Parallelism,
				Completions:  job.Spec.Completions,
				BackoffLimit: job.Spec.BackoffLimit,
			},
			Status: &pbSchedulerV1.BatchJobStatus{
				Active:    job.Status.Active,
				Succeeded: job.Status.Succeeded,
				Failed:    job.Status.Failed,
			},
		},
	}, nil
}

func (i *implementation) DeleteBatchJob(ctx context.Context, request *pbSchedulerV1.DeleteBatchJobRequest) (*pbSchedulerV1.DeleteBatchJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	var d meta.DeleteOptions

	if v := request.DeleteChildPods; v != nil {
		if *v {
			d.PropagationPolicy = util.NewType(meta.DeletePropagationBackground)
		} else {
			d.PropagationPolicy = util.NewType(meta.DeletePropagationOrphan)
		}
	}

	err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerBatchJobs(i.cfg.Namespace).Delete(ctx, request.GetName(), d)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.DeleteBatchJobResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	return &pbSchedulerV1.DeleteBatchJobResponse{Exists: true}, nil
}

func (i *implementation) ListBatchJob(ctx context.Context, request *pbSchedulerV1.ListBatchJobRequest) (*pbSchedulerV1.ListBatchJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	objects, err := list.ListObjects[*schedulerApi.ArangoSchedulerBatchJobList, *schedulerApi.ArangoSchedulerBatchJob](ctx, i.client.Arango().SchedulerV1beta1().ArangoSchedulerBatchJobs(i.cfg.Namespace), func(result *schedulerApi.ArangoSchedulerBatchJobList) []*schedulerApi.ArangoSchedulerBatchJob {
		r := make([]*schedulerApi.ArangoSchedulerBatchJob, len(result.Items))

		for id := range result.Items {
			r[id] = result.Items[id].DeepCopy()
		}

		return r
	})

	if err != nil {
		return nil, err
	}

	return &pbSchedulerV1.ListBatchJobResponse{
		BatchJobs: util.FormatList(objects, func(in *schedulerApi.ArangoSchedulerBatchJob) string {
			return in.GetName()
		}),
	}, nil
}
