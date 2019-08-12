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
	"reflect"
	"testing"

	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	fakeClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func newFakeHandler() *handler {
	f := fakeClientSet.NewSimpleClientset()

	return &handler{
		client: f,
	}
}

func newItem(operation operator.Operation, namespace, name string) operator.Item {
	return operator.Item{
		Group:   database.SchemeGroupVersion.Group,
		Version: database.SchemeGroupVersion.Version,
		Kind:    database.ArangoBackupPolicyResourceKind,

		Operation: operation,

		Namespace: namespace,
		Name:      name,
	}
}

func newItemFromBackupPolicy(operation operator.Operation, policy *database.ArangoBackupPolicy) operator.Item {
	return newItem(operation, policy.Namespace, policy.Name)
}

func newArangoBackupPolicy(schedule, namespace, name string, selector map[string]string, template database.ArangoBackupSpec) *database.ArangoBackupPolicy {
	return &database.ArangoBackupPolicy{
		TypeMeta: meta.TypeMeta{
			APIVersion: database.SchemeGroupVersion.String(),
			Kind:       database.ArangoBackupPolicyResourceKind,
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			SelfLink: fmt.Sprintf("/api/%s/%s/%s/%s",
				database.SchemeGroupVersion.String(),
				database.ArangoBackupPolicyResourcePlural,
				namespace,
				name),
			UID: uuid.NewUUID(),
		},
		Spec: database.ArangoBackupPolicySpec{
			Schedule: schedule,
			DeploymentSelector: &meta.LabelSelector{
				MatchLabels: selector,
			},
			BackupTemplate: template,
		},
	}
}

func refreshArangoBackupPolicy(t *testing.T, h *handler, policy *database.ArangoBackupPolicy) *database.ArangoBackupPolicy {
	newPolicy, err := h.client.DatabaseV1alpha().ArangoBackupPolicies(policy.Namespace).Get(policy.Name, meta.GetOptions{})
	require.NoError(t, err)

	return newPolicy
}

func createArangoBackupPolicy(t *testing.T, h *handler, policies ...*database.ArangoBackupPolicy) {
	for _, policy := range policies {
		_, err := h.client.DatabaseV1alpha().ArangoBackupPolicies(policy.Namespace).Create(policy)
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

func listArangoBackups(t *testing.T, handler *handler, namespace string) []database.ArangoBackup {
	result, err := handler.client.DatabaseV1alpha().ArangoBackups(namespace).List(meta.ListOptions{})
	require.NoError(t, err)

	return result.Items
}

func isInList(t *testing.T, backups []database.ArangoBackup, deployment *database.ArangoDeployment) {
	for _, backup := range backups {
		if backup.Spec.Deployment.Name == deployment.Name {
			return
		}
	}
	require.Fail(t, "backup is not present on list")
}

func hasOwnerReference(references []meta.OwnerReference, expectedReference meta.OwnerReference) bool {
	for _, ownerRef := range references {
		if reflect.DeepEqual(ownerRef, expectedReference) {
			return true
		}
	}

	return false
}
