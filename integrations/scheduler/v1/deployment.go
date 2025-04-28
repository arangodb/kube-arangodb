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

func (i *implementation) CreateDeployment(ctx context.Context, request *pbSchedulerV1.CreateDeploymentRequest) (*pbSchedulerV1.CreateDeploymentResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	template := scheduler.SpecAsTemplate(request.GetSpec())

	var spec schedulerApi.ArangoSchedulerDeployment

	spec.Namespace = i.cfg.Namespace

	if meta := request.GetSpec().GetMetadata(); meta != nil {
		if util.TypeOrDefault(meta.GenerateName, false) {
			spec.GenerateName = meta.Name
		} else {
			spec.Name = meta.Name
		}
	}

	spec.Spec.Template = *template

	if deployment := request.GetDeployment(); deployment != nil {
		spec.Spec.Replicas = deployment.Replicas
	}

	if jobSpec := request.GetSpec(); jobSpec != nil {
		if base := jobSpec.Base; base != nil {
			spec.Labels = base.Labels
			spec.Spec.Template.Labels = base.Labels
		}
	}

	job, err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerDeployments(i.cfg.Namespace).Create(ctx, &spec, meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	return &pbSchedulerV1.CreateDeploymentResponse{
		Name: job.Name,
	}, nil
}

func (i *implementation) GetDeployment(ctx context.Context, request *pbSchedulerV1.GetDeploymentRequest) (*pbSchedulerV1.GetDeploymentResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	deployment, err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerDeployments(i.cfg.Namespace).Get(ctx, request.GetName(), meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.GetDeploymentResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	return &pbSchedulerV1.GetDeploymentResponse{
		Exists: true,

		Deployment: &pbSchedulerV1.Deployment{
			Metadata: ExtractStatusMetadata(deployment.Status.ArangoSchedulerStatusMetadata),
			Spec: &pbSchedulerV1.DeploymentSpec{
				Replicas: deployment.Spec.Replicas,
			},
			Status: &pbSchedulerV1.DeploymentStatus{
				Replicas:            deployment.Status.DeploymentStatus.Replicas,
				UpdatedReplicas:     deployment.Status.DeploymentStatus.UpdatedReplicas,
				ReadyReplicas:       deployment.Status.DeploymentStatus.ReadyReplicas,
				AvailableReplicas:   deployment.Status.DeploymentStatus.AvailableReplicas,
				UnavailableReplicas: deployment.Status.DeploymentStatus.UnavailableReplicas,
			},
		},
	}, nil
}

func (i *implementation) UpdateDeployment(ctx context.Context, request *pbSchedulerV1.UpdateDeploymentRequest) (*pbSchedulerV1.UpdateDeploymentResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	job, err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerDeployments(i.cfg.Namespace).Get(ctx, request.GetName(), meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.UpdateDeploymentResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	if deployment := request.GetSpec(); deployment != nil {
		job.Spec.Replicas = deployment.Replicas
	}

	job, err = i.client.Arango().SchedulerV1beta1().ArangoSchedulerDeployments(i.cfg.Namespace).Update(ctx, job, meta.UpdateOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.UpdateDeploymentResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	return &pbSchedulerV1.UpdateDeploymentResponse{
		Exists: true,

		Deployment: &pbSchedulerV1.Deployment{
			Spec: &pbSchedulerV1.DeploymentSpec{
				Replicas: job.Spec.Replicas,
			},
		},
	}, nil
}

func (i *implementation) ListDeployment(ctx context.Context, request *pbSchedulerV1.ListDeploymentRequest) (*pbSchedulerV1.ListDeploymentResponse, error) {
	if request == nil {
		return nil, errors.Errorf("Request is nil")
	}

	objects, err := list.ListObjects[*schedulerApi.ArangoSchedulerDeploymentList, *schedulerApi.ArangoSchedulerDeployment](ctx, i.client.Arango().SchedulerV1beta1().ArangoSchedulerDeployments(i.cfg.Namespace), func(result *schedulerApi.ArangoSchedulerDeploymentList) []*schedulerApi.ArangoSchedulerDeployment {
		r := make([]*schedulerApi.ArangoSchedulerDeployment, len(result.Items))

		for id := range result.Items {
			r[id] = result.Items[id].DeepCopy()
		}

		return r
	})

	if err != nil {
		return nil, err
	}

	return &pbSchedulerV1.ListDeploymentResponse{
		Deployments: util.FormatList(objects, func(in *schedulerApi.ArangoSchedulerDeployment) string {
			return in.GetName()
		}),
	}, nil
}

func (i *implementation) DeleteDeployment(ctx context.Context, request *pbSchedulerV1.DeleteDeploymentRequest) (*pbSchedulerV1.DeleteDeploymentResponse, error) {
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

	err := i.client.Arango().SchedulerV1beta1().ArangoSchedulerDeployments(i.cfg.Namespace).Delete(ctx, request.GetName(), d)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return &pbSchedulerV1.DeleteDeploymentResponse{
				Exists: false,
			}, nil
		}

		return nil, err
	}

	return &pbSchedulerV1.DeleteDeploymentResponse{Exists: true}, nil
}
