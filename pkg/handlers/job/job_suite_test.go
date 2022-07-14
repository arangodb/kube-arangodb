//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	"testing"

	"github.com/stretchr/testify/require"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	appsApi "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	fakeClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func newFakeHandler() *handler {
	f := fakeClientSet.NewSimpleClientset()
	k := fake.NewSimpleClientset()

	h := &handler{
		client:        f,
		kubeClient:    k,
		eventRecorder: newEventInstance(event.NewEventRecorder("mock", k)),
		operator:      operator.NewOperator("mock", "mock", "mock"),
	}

	return h
}

func newItem(o operation.Operation, namespace, name string) operation.Item {
	return operation.Item{
		Group:   appsApi.SchemeGroupVersion.Group,
		Version: appsApi.SchemeGroupVersion.Version,
		Kind:    apps.ArangoJobResourceKind,

		Operation: o,

		Namespace: namespace,
		Name:      name,
	}
}

func newItemFromJob(operation operation.Operation, job *appsApi.ArangoJob) operation.Item { // nolint:unparam
	return newItem(operation, job.Namespace, job.Name)
}

func refreshArangoJob(t *testing.T, h *handler, job *appsApi.ArangoJob) *appsApi.ArangoJob {
	newJob, err := h.client.AppsV1().ArangoJobs(job.Namespace).Get(context.Background(), job.Name, meta.GetOptions{})
	require.NoError(t, err)

	return newJob
}

func createArangoJob(t *testing.T, h *handler, jobs ...*appsApi.ArangoJob) {
	for _, job := range jobs {
		_, err := h.client.AppsV1().ArangoJobs(job.Namespace).Create(context.Background(), job, meta.CreateOptions{})
		require.NoError(t, err)
	}
}

func createK8sJob(t *testing.T, h *handler, jobs ...*batch.Job) {
	for _, job := range jobs {
		_, err := h.kubeClient.BatchV1().Jobs(job.Namespace).Create(context.Background(), job, meta.CreateOptions{})
		require.NoError(t, err)
	}
}

func createArangoDeployment(t *testing.T, h *handler, deployments ...*deploymentApi.ArangoDeployment) {
	for _, deployment := range deployments {
		_, err := h.client.DatabaseV1().ArangoDeployments(deployment.Namespace).Create(context.Background(), deployment, meta.CreateOptions{})
		require.NoError(t, err)
	}
}

func newArangoJob(name, namespace, deployment string) *appsApi.ArangoJob {
	return &appsApi.ArangoJob{
		TypeMeta: meta.TypeMeta{
			APIVersion: appsApi.SchemeGroupVersion.String(),
			Kind:       apps.ArangoJobResourceKind,
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       uuid.NewUUID(),
		},
		Spec: appsApi.ArangoJobSpec{
			ArangoDeploymentName: deployment,
			JobTemplate: &batch.JobSpec{
				Template: core.PodTemplateSpec{
					Spec: core.PodSpec{
						Containers: []core.Container{
							{
								Image: "perl",
								Name:  "pi",
								Args:  []string{"perl", "-Mbignum=bpi", "-wle", "print bpi(2000)"},
							},
						},
						RestartPolicy: core.RestartPolicyNever,
					},
				},
			},
		},
	}
}

func newArangoDeployment(name, namespace string) *deploymentApi.ArangoDeployment {
	return &deploymentApi.ArangoDeployment{
		TypeMeta: meta.TypeMeta{
			APIVersion: deploymentApi.SchemeGroupVersion.String(),
			Kind:       deployment.ArangoDeploymentResourceKind,
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       uuid.NewUUID(),
		},
	}
}

func newK8sJob(name, namespace string) *batch.Job {
	return &batch.Job{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       uuid.NewUUID(),
		},
	}
}
