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

	"google.golang.org/grpc"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/scheduler"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

var _ pbSchedulerV1.SchedulerV1Server = &implementation{}
var _ svc.Handler = &implementation{}

func New(ctx context.Context, client kclient.Client, cfg Configuration) (svc.Handler, error) {
	return newInternal(ctx, client, cfg)
}

func newInternal(ctx context.Context, client kclient.Client, cfg Configuration) (*implementation, error) {
	if cfg.VerifyAccess {
		// Lets Verify Access
		if status := kresources.VerifyAccessRequestStatus(ctx, client.Kubernetes(), "create", "batch", "v1", "jobs", "", "", cfg.Namespace); !status.Allowed {
			return nil, errors.Errorf("Unable to access Jobs BatchV1 Create: %s", status.Reason)
		}
		if status := kresources.VerifyAccessRequestStatus(ctx, client.Kubernetes(), "delete", "batch", "v1", "jobs", "", "", cfg.Namespace); !status.Allowed {
			return nil, errors.Errorf("Unable to access Jobs BatchV1 Create: %s", status.Reason)
		}
		if status := kresources.VerifyAccessRequestStatus(ctx, client.Kubernetes(), "get", "batch", "v1", "jobs", "", "", cfg.Namespace); !status.Allowed {
			return nil, errors.Errorf("Unable to access Jobs BatchV1 Create: %s", status.Reason)
		}
	}

	return &implementation{
		cfg:       cfg,
		client:    client,
		scheduler: scheduler.NewScheduler(client, cfg.Namespace),
	}, nil
}

type implementation struct {
	cfg Configuration

	client    kclient.Client
	scheduler scheduler.Scheduler

	pbSchedulerV1.UnimplementedSchedulerV1Server
}

func (i *implementation) Name() string {
	return pbSchedulerV1.Name
}

func (i *implementation) Register(registrar *grpc.Server) {
	pbSchedulerV1.RegisterSchedulerV1Server(registrar, i)
}

func (i *implementation) Health() svc.HealthState {
	return svc.Healthy
}

func (i *implementation) CreateBatchJob(ctx context.Context, request *pbSchedulerV1.CreateBatchJobRequest) (*pbSchedulerV1.CreateBatchJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	rendered, profiles, err := i.scheduler.Render(ctx, request.GetSpec())
	if err != nil {
		return nil, err
	}

	rendered.Spec.RestartPolicy = core.RestartPolicyNever

	var spec batch.Job

	spec.Namespace = i.cfg.Namespace

	if meta := request.GetSpec().GetMetadata(); meta != nil {
		if util.TypeOrDefault(meta.GenerateName, false) {
			spec.GenerateName = meta.Name
		} else {
			spec.Name = meta.Name
		}
	}

	spec.Spec.Template = *rendered

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
		if job := batchJobSpec.Job; job != nil {
			spec.Labels = job.Labels
		}
	}

	job, err := i.client.Kubernetes().BatchV1().Jobs(i.cfg.Namespace).Create(ctx, &spec, meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	return &pbSchedulerV1.CreateBatchJobResponse{
		Name:     job.Name,
		Profiles: profiles,
	}, nil
}

func (i *implementation) GetBatchJob(ctx context.Context, request *pbSchedulerV1.GetBatchJobRequest) (*pbSchedulerV1.GetBatchJobResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	job, err := i.client.Kubernetes().BatchV1().Jobs(i.cfg.Namespace).Get(ctx, request.GetName(), meta.GetOptions{})
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

	err := i.client.Kubernetes().BatchV1().Jobs(i.cfg.Namespace).Delete(ctx, request.GetName(), d)
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
