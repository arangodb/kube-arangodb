//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package policy

import (
	"fmt"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/event"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"

	"k8s.io/client-go/kubernetes/fake"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	fakeClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func newFakeHandler() *handler {
	f := fakeClientSet.NewSimpleClientset()
	k := fake.NewSimpleClientset()

	h := &handler{
		client:        f,
		kubeClient:    k,
		eventRecorder: newEventInstance(event.NewEventRecorder("mock", k)),
	}

	return h
}

func newItem(o operation.Operation, namespace, name string) operation.Item {
	return operation.Item{
		Group:   database.SchemeGroupVersion.Group,
		Version: database.SchemeGroupVersion.Version,
		Kind:    backupApi.ArangoBackupPolicyResourceKind,

		Operation: o,

		Namespace: namespace,
		Name:      name,
	}
}

func newItemFromBackupPolicy(operation operation.Operation, policy *backupApi.ArangoBackupPolicy) operation.Item {
	return newItem(operation, policy.Namespace, policy.Name)
}

func newArangoBackupPolicy(schedule, namespace, name string, selector map[string]string, template backupApi.ArangoBackupTemplate) *backupApi.ArangoBackupPolicy {
	return &backupApi.ArangoBackupPolicy{
		TypeMeta: meta.TypeMeta{
			APIVersion: database.SchemeGroupVersion.String(),
			Kind:       backupApi.ArangoBackupPolicyResourceKind,
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			SelfLink: fmt.Sprintf("/api/%s/%s/%s/%s",
				database.SchemeGroupVersion.String(),
				backupApi.ArangoBackupPolicyResourcePlural,
				namespace,
				name),
			UID: uuid.NewUUID(),
		},
		Spec: backupApi.ArangoBackupPolicySpec{
			Schedule: schedule,
			DeploymentSelector: &meta.LabelSelector{
				MatchLabels: selector,
			},
			BackupTemplate: template,
		},
	}
}

func refreshArangoBackupPolicy(t *testing.T, h *handler, policy *backupApi.ArangoBackupPolicy) *backupApi.ArangoBackupPolicy {
	newPolicy, err := h.client.BackupV1alpha().ArangoBackupPolicies(policy.Namespace).Get(policy.Name, meta.GetOptions{})
	require.NoError(t, err)

	return newPolicy
}

func createArangoBackupPolicy(t *testing.T, h *handler, policies ...*backupApi.ArangoBackupPolicy) {
	for _, policy := range policies {
		_, err := h.client.BackupV1alpha().ArangoBackupPolicies(policy.Namespace).Create(policy)
		require.NoError(t, err)
	}
}

func newArangoDeployment(namespace string, labels map[string]string) *database.ArangoDeployment {
	name := string(uuid.NewUUID())
	return &database.ArangoDeployment{
		TypeMeta: meta.TypeMeta{
			APIVersion: database.SchemeGroupVersion.String(),
			Kind:       database.ArangoDeploymentResourceKind,
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			SelfLink: fmt.Sprintf("/api/%s/%s/%s/%s",
				database.SchemeGroupVersion.String(),
				database.ArangoDeploymentResourcePlural,
				namespace,
				name),
			UID:    uuid.NewUUID(),
			Labels: labels,
		},
	}
}

func createArangoDeployment(t *testing.T, h *handler, deployments ...*database.ArangoDeployment) {
	for _, deployment := range deployments {
		_, err := h.client.DatabaseV1alpha().ArangoDeployments(deployment.Namespace).Create(deployment)
		require.NoError(t, err)
	}
}

func listArangoBackups(t *testing.T, handler *handler, namespace string) []backupApi.ArangoBackup {
	result, err := handler.client.BackupV1alpha().ArangoBackups(namespace).List(meta.ListOptions{})
	require.NoError(t, err)

	return result.Items
}

func isInList(t *testing.T, backups []backupApi.ArangoBackup, deployment *database.ArangoDeployment) {
	for _, backup := range backups {
		if backup.Spec.Deployment.Name == deployment.Name {
			return
		}
	}
	require.Fail(t, "backup is not present on list")
}
