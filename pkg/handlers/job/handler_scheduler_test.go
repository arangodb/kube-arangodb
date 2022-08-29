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
	"testing"

	"github.com/stretchr/testify/require"
	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func Test_Job_Create(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())
	deployment := string(uuid.NewUUID())

	job := newArangoJob(name, namespace, deployment)
	database := newArangoDeployment(deployment, namespace)

	// Act
	createArangoJob(t, handler, job)
	createArangoDeployment(t, handler, database)
	require.NoError(t, handler.Handle(newItemFromJob(operation.Add, job)))

	// Assert
	newJob := refreshArangoJob(t, handler, job)
	require.Empty(t, newJob.Status.Conditions)
	require.True(t, len(newJob.Spec.JobTemplate.Template.Spec.Containers) == 1)
	require.True(t, newJob.Spec.JobTemplate.Template.Spec.Containers[0].Image == job.Spec.JobTemplate.Template.Spec.Containers[0].Image)
}

func Test_Job_Update(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())
	deployment := string(uuid.NewUUID())

	job := newArangoJob(name, namespace, deployment)
	k8sJob := newK8sJob(name, namespace)

	// Act
	createArangoJob(t, handler, job)
	createK8sJob(t, handler, k8sJob)
	require.NoError(t, handler.Handle(newItemFromJob(operation.Update, job)))

	// Assert
	newJob := refreshArangoJob(t, handler, job)
	require.Empty(t, newJob.Status.Conditions)
	require.True(t, len(newJob.Spec.JobTemplate.Template.Spec.Containers) == 1)
	require.True(t, newJob.Spec.JobTemplate.Template.Spec.Containers[0].Image == job.Spec.JobTemplate.Template.Spec.Containers[0].Image)
}

func Test_Job_Create_Error_If_Deployment_Not_Exist(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())
	deployment := string(uuid.NewUUID())

	job := newArangoJob(name, namespace, deployment)

	// Act
	createArangoJob(t, handler, job)
	require.NoError(t, handler.Handle(newItemFromJob(operation.Update, job)))

	// Assert
	newJob := refreshArangoJob(t, handler, job)
	require.True(t, len(newJob.Status.Conditions) == 1)
	require.True(t, newJob.Status.Conditions[0].Type == batch.JobFailed)
}
