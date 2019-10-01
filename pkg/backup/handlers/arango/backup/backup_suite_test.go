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

package backup

import (
	"fmt"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/event"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"
	"k8s.io/client-go/kubernetes/fake"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/state"
	fakeClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	errorString = "errorString"
)

func newFakeHandler() *handler {
	f := fakeClientSet.NewSimpleClientset()
	k := fake.NewSimpleClientset()

	return &handler{
		client:     f,
		kubeClient: k,

		arangoClientTimeout: defaultArangoClientTimeout,
		eventRecorder:       newEventInstance(event.NewEventRecorder("mock", k)),
	}
}

func newErrorsFakeHandler(errors mockErrorsArangoClientBackup) (*handler, *mockArangoClientBackup) {
	handler := newFakeHandler()

	mock := newMockArangoClientBackup(errors)
	handler.arangoClientFactory = newMockArangoClientBackupFactory(mock)

	return handler, mock
}

func newObjectSet(state state.State) (*backupApi.ArangoBackup, *database.ArangoDeployment) {
	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())

	obj := newArangoBackup(name, namespace, name, state)
	deployment := newArangoDeployment(namespace, name)

	return obj, deployment
}

func compareTemporaryState(t *testing.T, err error, errorMsg string, handler *handler, obj *backupApi.ArangoBackup) {
	require.Error(t, err)
	require.True(t, IsTemporaryError(err))
	require.EqualError(t, err, errorMsg)

	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, obj.Status, newObj.Status)
}

func newItem(o operation.Operation, namespace, name string) operation.Item {
	return operation.Item{
		Group:   database.SchemeGroupVersion.Group,
		Version: database.SchemeGroupVersion.Version,
		Kind:    backupApi.ArangoBackupResourceKind,

		Operation: o,

		Namespace: namespace,
		Name:      name,
	}
}

func newItemFromBackup(operation operation.Operation, backup *backupApi.ArangoBackup) operation.Item {
	return newItem(operation, backup.Namespace, backup.Name)
}

func newArangoBackup(objectRef, namespace, name string, state state.State) *backupApi.ArangoBackup {
	return &backupApi.ArangoBackup{
		TypeMeta: meta.TypeMeta{
			APIVersion: database.SchemeGroupVersion.String(),
			Kind:       backupApi.ArangoBackupResourceKind,
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			SelfLink: fmt.Sprintf("/api/%s/%s/%s/%s",
				database.SchemeGroupVersion.String(),
				backupApi.ArangoBackupResourcePlural,
				namespace,
				name),
			UID:        uuid.NewUUID(),
			Finalizers: backupApi.FinalizersArangoBackup,
		},
		Spec: backupApi.ArangoBackupSpec{
			Deployment: backupApi.ArangoBackupSpecDeployment{
				Name: objectRef,
			},
		},
		Status: backupApi.ArangoBackupStatus{
			ArangoBackupState: backupApi.ArangoBackupState{
				State: state,
			},
		},
	}
}

func createArangoBackup(t *testing.T, h *handler, backups ...*backupApi.ArangoBackup) {
	for _, backup := range backups {
		_, err := h.client.BackupV1alpha().ArangoBackups(backup.Namespace).Create(backup)
		require.NoError(t, err)
	}
}

func refreshArangoBackup(t *testing.T, h *handler, backup *backupApi.ArangoBackup) *backupApi.ArangoBackup {
	obj, err := h.client.BackupV1alpha().ArangoBackups(backup.Namespace).Get(backup.Name, meta.GetOptions{})
	require.NoError(t, err)
	return obj
}

func newArangoDeployment(namespace, name string) *database.ArangoDeployment {
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
			UID: uuid.NewUUID(),
		},
	}
}

func createArangoDeployment(t *testing.T, h *handler, deployments ...*database.ArangoDeployment) {
	for _, deployment := range deployments {
		_, err := h.client.DatabaseV1alpha().ArangoDeployments(deployment.Namespace).Create(deployment)
		require.NoError(t, err)
	}
}

func wrapperUndefinedDeployment(t *testing.T, state state.State) {
	t.Run("Empty Name", func(t *testing.T) {
		// Arrange
		handler := newFakeHandler()

		obj, _ := newObjectSet(state)
		obj.Spec.Deployment.Name = ""

		// Act
		createArangoBackup(t, handler, obj)
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateFailed)

		require.Equal(t, newObj.Status.Message, createFailMessage(state, "deployment name can not be empty"))
	})

	t.Run("Missing Deployment", func(t *testing.T) {
		// Arrange
		handler := newFakeHandler()

		obj, _ := newObjectSet(state)

		// Act
		createArangoBackup(t, handler, obj)
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateFailed)

		require.Equal(t, newObj.Status.Message, createFailMessage(state, fmt.Sprintf("%s \"%s\" not found", database.ArangoDeploymentCRDName, obj.Name)))
	})
}

func wrapperConnectionIssues(t *testing.T, state state.State) {
	t.Run("Unable to create deployment client", func(t *testing.T) {
		// Arrange
		handler := newFakeHandler()

		f := newMockArangoClientBackupErrorFactory(fmt.Errorf(errorString))
		handler.arangoClientFactory = f

		obj, deployment := newObjectSet(state)

		// Act
		createArangoBackup(t, handler, obj)
		createArangoDeployment(t, handler, deployment)
		err := handler.Handle(newItemFromBackup(operation.Update, obj))

		// Assert
		require.Error(t, err)
		require.True(t, IsTemporaryError(err))

		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, newObj.Status.State, state)

	})
}

func wrapperProgressMissing(t *testing.T, state state.State) {
	t.Run("Backup Progress Missing", func(t *testing.T) {
		// Arrange
		handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

		obj, deployment := newObjectSet(state)

		// Act
		createArangoBackup(t, handler, obj)
		createArangoDeployment(t, handler, deployment)
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateFailed)

		require.Equal(t, newObj.Status.Message, createFailMessage(state, fmt.Sprintf("backup details are missing")))

	})
}
