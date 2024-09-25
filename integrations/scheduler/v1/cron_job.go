//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/generators/kubernetes"
	"github.com/arangodb/kube-arangodb/pkg/scheduler"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func (i *implementation) CreateCronJob(ctx context.Context, request *pbSchedulerV1.CreateCronJobRequest) (*pbSchedulerV1.CreateCronJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	template := scheduler.SpecAsTemplate(request.GetSpec())

	var spec schedulerApi.ArangoSchedulerCronJob

	spec.Namespace = i.cfg.Namespace

	if meta := request.GetSpec().GetMetadata(); meta != nil {
		if util.TypeOrDefault(meta.GenerateName, false) {
			spec.GenerateName = meta.Name
		} else {
			spec.Name = meta.Name
		}
	}

	spec.Spec.JobTemplate.Spec.Template = *template

	if cronJob := request.GetCronJob(); cronJob != nil {
		spec.Spec.Schedule = cronJob.Schedule

		if batchJob := cronJob.GetJob(); batchJob != nil {
			if v := batchJob.Completions; v != nil {
				spec.Spec.JobTemplate.Spec.Completions = v
			}

			if v := batchJob.Parallelism; v != nil {
				spec.Spec.JobTemplate.Spec.Parallelism = v
			}

			if v := batchJob.BackoffLimit; v != nil {
				spec.Spec.JobTemplate.Spec.BackoffLimit = v
			}
		}
	}

	if batchJobSpec := request.GetSpec(); batchJobSpec != nil {
		if base := batchJobSpec.Base; base != nil {
			spec.Labels = base.Labels
			spec.Spec.JobTemplate.Labels = base.Labels
		}
	}

	job, err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerCronJobs(i.cfg.Namespace).Create(ctx, &spec, meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	return &pbSchedulerV1.CreateCronJobResponse{
		Name: job.Name,
	}, nil
}

func (i *implementation) GetCronJob(ctx context.Context, request *pbSchedulerV1.GetCronJobRequest) (*pbSchedulerV1.GetCronJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	job, err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerCronJobs(i.cfg.Namespace).Get(ctx, request.GetName(), meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.GetCronJobResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	return &pbSchedulerV1.GetCronJobResponse{
		Exists: true,

		CronJob: &pbSchedulerV1.CronJob{
			Metadata: ExtractStatusMetadata(job.Status.ArangoSchedulerStatusMetadata),
			Spec: &pbSchedulerV1.CronJobSpec{
				Schedule: job.Spec.Schedule,

				Job: &pbSchedulerV1.BatchJobSpec{
					Parallelism:  job.Spec.JobTemplate.Spec.Parallelism,
					Completions:  job.Spec.JobTemplate.Spec.Completions,
					BackoffLimit: job.Spec.JobTemplate.Spec.BackoffLimit,
				},
			},
			Status: &pbSchedulerV1.CronJobStatus{
				BatchJobs: util.FormatList(job.Status.Active, func(in core.ObjectReference) string {
					return in.Name
				}),
			},
		},
	}, nil
}

func (i *implementation) UpdateCronJob(ctx context.Context, request *pbSchedulerV1.UpdateCronJobRequest) (*pbSchedulerV1.UpdateCronJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	job, err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerCronJobs(i.cfg.Namespace).Get(ctx, request.GetName(), meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.UpdateCronJobResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	if cronJob := request.GetSpec(); cronJob != nil {
		job.Spec.Schedule = cronJob.Schedule

		if batchJob := cronJob.GetJob(); batchJob != nil {
			if v := batchJob.Completions; v != nil {
				job.Spec.JobTemplate.Spec.Completions = v
			}

			if v := batchJob.Parallelism; v != nil {
				job.Spec.JobTemplate.Spec.Parallelism = v
			}

			if v := batchJob.BackoffLimit; v != nil {
				job.Spec.JobTemplate.Spec.BackoffLimit = v
			}
		}
	}

	job, err = i.client.Arango().SchedulerV1beta1().ArangoSchedulerCronJobs(i.cfg.Namespace).Update(ctx, job, meta.UpdateOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.UpdateCronJobResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	return &pbSchedulerV1.UpdateCronJobResponse{
		Exists: true,

		CronJob: &pbSchedulerV1.CronJob{
			Spec: &pbSchedulerV1.CronJobSpec{
				Schedule: job.Spec.Schedule,

				Job: &pbSchedulerV1.BatchJobSpec{
					Parallelism:  job.Spec.JobTemplate.Spec.Parallelism,
					Completions:  job.Spec.JobTemplate.Spec.Completions,
					BackoffLimit: job.Spec.JobTemplate.Spec.BackoffLimit,
				},
			},
		},
	}, nil
}

func (i *implementation) ListCronJob(ctx context.Context, request *pbSchedulerV1.ListCronJobRequest) (*pbSchedulerV1.ListCronJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	objects, err := kubernetes.ListObjects[*schedulerApi.ArangoSchedulerCronJobList, *schedulerApi.ArangoSchedulerCronJob](ctx, i.client.Arango().SchedulerV1beta1().ArangoSchedulerCronJobs(i.cfg.Namespace), func(result *schedulerApi.ArangoSchedulerCronJobList) []*schedulerApi.ArangoSchedulerCronJob {
		r := make([]*schedulerApi.ArangoSchedulerCronJob, len(result.Items))

		for id := range result.Items {
			r[id] = result.Items[id].DeepCopy()
		}

		return r
	})

	if err != nil {
		return nil, err
	}

	return &pbSchedulerV1.ListCronJobResponse{
		CronJobs: util.FormatList(objects, func(in *schedulerApi.ArangoSchedulerCronJob) string {
			return in.GetName()
		}),
	}, nil
}

func (i *implementation) DeleteCronJob(ctx context.Context, request *pbSchedulerV1.DeleteCronJobRequest) (*pbSchedulerV1.DeleteCronJobResponse, error) {
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

	err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerCronJobs(i.cfg.Namespace).Delete(ctx, request.GetName(), d)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.DeleteCronJobResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	return &pbSchedulerV1.DeleteCronJobResponse{Exists: true}, nil
}
