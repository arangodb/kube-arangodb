//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package job

import (
	"context"
	"fmt"
	"os"
	"reflect"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	appsApi "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

const (
	jobCreatedUpdated = "ArangoJobCreatedOrUpdated"
	jobError          = "Error"
)

type handler struct {
	client        arangoClientSet.Interface
	kubeClient    kubernetes.Interface
	eventRecorder event.RecorderInstance

	operator operator.Operator
}

func (*handler) Name() string {
	return apps.ArangoJobResourceKind
}

func (h *handler) Handle(item operation.Item) error {
	// Do not act on delete event
	if item.Operation == operation.Delete {
		return nil
	}

	// Get Job object. It also covers NotFound case
	job, err := h.client.AppsV1().ArangoJobs(item.Namespace).Get(context.Background(), item.Name, meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		logger.Error("ArangoJob fetch error %v", err)
		return err
	}

	status := h.processArangoJob(job.DeepCopy())
	if reflect.DeepEqual(job.Status, status) {
		return nil
	}

	job.Status = status

	// Update status on object
	if _, err = h.client.AppsV1().ArangoJobs(item.Namespace).UpdateStatus(context.Background(), job, meta.UpdateOptions{}); err != nil {
		logger.Error("ArangoJob status update error %v", err)
		return err
	}

	return nil
}

func (h *handler) createFailedJobStatusWithEvent(msg string, job *appsApi.ArangoJob) batch.JobStatus {
	h.eventRecorder.Warning(job, jobError, msg)
	return batch.JobStatus{
		Conditions: []batch.JobCondition{
			{
				Type:    batch.JobFailed,
				Status:  core.ConditionUnknown,
				Message: msg,
			},
		},
	}
}

func (h *handler) processArangoJob(job *appsApi.ArangoJob) batch.JobStatus {
	existingJob, err := h.kubeClient.BatchV1().Jobs(job.Namespace).Get(context.Background(), job.Name, meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			k8sJob, err := h.prepareK8sJob(job)
			if err != nil {
				return h.createFailedJobStatusWithEvent(fmt.Sprintf("can not prepare k8s Job: %s", err.Error()), job)
			}

			existingJob, err = h.kubeClient.BatchV1().Jobs(job.Namespace).Create(context.Background(), k8sJob, meta.CreateOptions{})
			if err != nil {
				return h.createFailedJobStatusWithEvent(fmt.Sprintf("can not create k8s Job: %s", err.Error()), job)
			}
			h.eventRecorder.Normal(job, jobCreatedUpdated, "Arango job has been updated/created")
		} else {
			return h.createFailedJobStatusWithEvent(fmt.Sprintf("can not check if k8s Job alreadt exist: %s", err.Error()), job)
		}
	}

	return existingJob.Status
}

func (h *handler) prepareK8sJob(job *appsApi.ArangoJob) (*batch.Job, error) {
	k8sJob := batch.Job{}
	k8sJob.Name = job.Name
	k8sJob.Namespace = job.Namespace
	k8sJob.Spec = *job.Spec.JobTemplate
	k8sJob.Spec.Template.Spec.ServiceAccountName = os.Getenv(constants.EnvArangoJobSAName)
	k8sJob.SetOwnerReferences(append(job.GetOwnerReferences(), job.AsOwner()))

	deployment, err := h.client.DatabaseV1().ArangoDeployments(job.Namespace).Get(context.Background(), job.Spec.ArangoDeploymentName, meta.GetOptions{})
	if err != nil {
		logger.Error("ArangoDeployment fetch error %v", err)
		return &k8sJob, err
	}

	spec := deployment.GetAcceptedSpec()

	if spec.TLS.IsSecure() {
		k8sJob.Spec.Template.Spec.Volumes = []core.Volume{
			{
				Name: shared.TlsKeyfileVolumeName,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: spec.TLS.GetCASecretName(),
					},
				},
			},
		}
	}

	executable, err := os.Executable()
	if err != nil {
		logger.Error("reading Operator executable name error %v", err)
		return &k8sJob, err
	}

	initContainer := k8sutil.ArangodWaiterInitContainer(api.ServerGroupReservedInitContainerNameWait, deployment.Name, executable,
		h.operator.Image(), spec.TLS.IsSecure(), &core.SecurityContext{})

	k8sJob.Spec.Template.Spec.InitContainers = append(k8sJob.Spec.Template.Spec.InitContainers, initContainer)

	return &k8sJob, nil
}

func (*handler) CanBeHandled(item operation.Item) bool {
	return item.Group == appsApi.SchemeGroupVersion.Group &&
		item.Version == appsApi.SchemeGroupVersion.Version &&
		item.Kind == apps.ArangoJobResourceKind
}
