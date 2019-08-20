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
	"testing"
	"time"

	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Finalizer_PassThru(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, _ := newObjectSet(database.ArangoBackupStateCreate)
	time := meta.Time{
		Time: time.Now(),
	}
	obj.DeletionTimestamp = &time

	// Act
	//createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)
}

func Test_Finalizer_RemoveObject(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Finalizers = []string{
		database.FinalizerArangoBackup,
	}

	time := meta.Now()
	obj.DeletionTimestamp = &time

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Forced:            &backupMeta.Forced,
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)

	require.Len(t, newObj.Finalizers, 0)

	exists, err := mock.Exists(backupMeta.ID)
	require.NoError(t, err)
	require.False(t, exists)
}

func Test_Finalizer_RemoveObject_WithoutFinalizer(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)

	time := meta.Now()
	obj.DeletionTimestamp = &time

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Forced:            &backupMeta.Forced,
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Finalizers = nil

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)

	require.Len(t, newObj.Finalizers, 0)

	exists, err := mock.Exists(backupMeta.ID)
	require.NoError(t, err)
	require.True(t, exists)
}

func Test_Finalizer_RemoveObject_UnknownFinalizer(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Finalizers = []string{
		"UNKNOWN",
	}

	time := meta.Now()
	obj.DeletionTimestamp = &time

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Forced:            &backupMeta.Forced,
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)

	require.Len(t, newObj.Finalizers, 1)

	exists, err := mock.Exists(backupMeta.ID)
	require.NoError(t, err)
	require.True(t, exists)
}

func Test_Finalizer_RemoveObject_MixedFinalizers(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Finalizers = []string{
		"UNKNOWN",
		database.FinalizerArangoBackup,
	}

	time := meta.Now()
	obj.DeletionTimestamp = &time

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Forced:            &backupMeta.Forced,
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)

	require.Len(t, newObj.Finalizers, 1)

	exists, err := mock.Exists(backupMeta.ID)
	require.NoError(t, err)
	require.False(t, exists)
}

func Test_Finalizer_AddDefault(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateNone)

	obj.Finalizers = nil

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.NotNil(t, newObj.Finalizers)
	require.True(t, hasFinalizers(newObj))
}

func Test_Finalizer_AppendDefault(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateNone)

	obj.Finalizers = []string{
		"RANDOM",
		"FINALIZERS",
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.NotNil(t, newObj.Finalizers)
	require.True(t, hasFinalizers(newObj))
}
