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

package backup

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	fakeClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	"github.com/arangodb/kube-arangodb/pkg/handlers/backup/state"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
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

	return handler, &mockArangoClientBackup{
		backup: nil,
		state:  mock,
	}
}

func newObjectSet(state state.State) (*backupApi.ArangoBackup, *database.ArangoDeployment) {
	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())

	obj := newArangoBackup(name, namespace, name, state)
	deployment := newArangoDeployment(namespace, name)

	return obj, deployment
}

func newItem(o operation.Operation, namespace, name string) operation.Item {
	return operation.Item{
		Group:   backupApi.SchemeGroupVersion.Group,
		Version: backupApi.SchemeGroupVersion.Version,
		Kind:    backup.ArangoBackupResourceKind,

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
			APIVersion: backupApi.SchemeGroupVersion.String(),
			Kind:       backup.ArangoBackupResourceKind,
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			SelfLink: fmt.Sprintf("/api/%s/%s/%s/%s",
				backupApi.SchemeGroupVersion.String(),
				backup.ArangoBackupResourcePlural,
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
		_, err := h.client.BackupV1().ArangoBackups(backup.Namespace).Create(context.Background(), backup, meta.CreateOptions{})
		require.NoError(t, err)
	}
}

func refreshArangoBackup(t *testing.T, h *handler, backup *backupApi.ArangoBackup) *backupApi.ArangoBackup {
	obj, err := h.client.BackupV1().ArangoBackups(backup.Namespace).Get(context.Background(), backup.Name, meta.GetOptions{})
	require.NoError(t, err)
	return obj
}

func newArangoDeployment(namespace, name string) *database.ArangoDeployment {
	return &database.ArangoDeployment{
		TypeMeta: meta.TypeMeta{
			APIVersion: backupApi.SchemeGroupVersion.String(),
			Kind:       deployment.ArangoDeploymentResourceKind,
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			SelfLink: fmt.Sprintf("/api/%s/%s/%s/%s",
				backupApi.SchemeGroupVersion.String(),
				deployment.ArangoDeploymentResourcePlural,
				namespace,
				name),
			UID: uuid.NewUUID(),
		},
	}
}

func createArangoDeployment(t *testing.T, h *handler, deployments ...*database.ArangoDeployment) {
	for _, deployment := range deployments {
		_, err := h.client.DatabaseV1().ArangoDeployments(deployment.Namespace).Create(context.Background(), deployment, meta.CreateOptions{})
		require.NoError(t, err)
	}
}

func compareBackupMeta(t *testing.T, backupMeta driver.BackupMeta, backup *backupApi.ArangoBackup) {
	require.NotNil(t, backup.Status.Backup)
	require.Equal(t, string(backupMeta.ID), backup.Status.Backup.ID)
	require.Equal(t, backupMeta.PotentiallyInconsistent, *backup.Status.Backup.PotentiallyInconsistent)
	require.Equal(t, backupMeta.SizeInBytes, backup.Status.Backup.SizeInBytes)
	require.Equal(t, backupMeta.DateTime.UTC().Unix(), backup.Status.Backup.CreationTimestamp.Time.UTC().Unix())
	require.Equal(t, backupMeta.NumberOfDBServers, backup.Status.Backup.NumberOfDBServers)
	require.Equal(t, backupMeta.Version, backup.Status.Backup.Version)
}

func checkBackup(t *testing.T, backup *backupApi.ArangoBackup, state state.State, available bool) {
	require.Equal(t, state, backup.Status.State)
	require.Equal(t, available, backup.Status.Available)
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

		require.Equal(t, newObj.Status.Message, createStateMessage(state, backupApi.ArangoBackupStateFailed, "deployment name can not be empty"))
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

		require.Equal(t, newObj.Status.Message, createStateMessage(state, backupApi.ArangoBackupStateFailed, fmt.Sprintf("%s \"%s\" not found", deployment.ArangoDeploymentCRDName, obj.Name)))
	})
}

func wrapperConnectionIssues(t *testing.T, state state.State) {
	t.Run("Unable to create deployment client", func(t *testing.T) {
		// Arrange
		handler := newFakeHandler()

		f := newMockArangoClientBackupErrorFactory(errors.Newf(errorString))
		handler.arangoClientFactory = f

		obj, deployment := newObjectSet(state)

		// Act
		createArangoBackup(t, handler, obj)
		createArangoDeployment(t, handler, deployment)
		err := handler.Handle(newItemFromBackup(operation.Update, obj))

		// Assert
		require.Error(t, err)
		require.True(t, isTemporaryError(err))

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

		require.Equal(t, newObj.Status.Message, createStateMessage(state, backupApi.ArangoBackupStateFailed, "missing field .status.backup"))
	})
}
