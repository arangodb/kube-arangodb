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

	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_State_Ready_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, database.ArangoBackupStateReady)
	wrapperConnectionIssues(t, database.ArangoBackupStateReady)
}

func Test_State_Ready_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, obj.Status, newObj.Status)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_GetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateDeleted, newObj.Status.State)
	require.False(t, newObj.Status.Available)
}

func Test_State_Ready_TemporaryGetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError: true,
		getError:         errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err = handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}

func Test_State_Ready_Upload(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Spec.Upload = &database.ArangoBackupSpecOperation{
		RepositoryPath: "Any",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateUpload, newObj.Status.State)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_DownloadDoNothing(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Spec.Download = &database.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: database.ArangoBackupSpecOperation{
			RepositoryPath: "any",
		},
		ID: "some",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateReady, newObj.Status.State)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_DoNotUploadDownloadedBackup(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Spec.Upload = &database.ArangoBackupSpecOperation{
		RepositoryPath: "Any",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	trueVar := true

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Downloaded:        &trueVar,
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateReady, newObj.Status.State)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_DoNotReUploadBackup(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Spec.Upload = &database.ArangoBackupSpecOperation{
		RepositoryPath: "Any",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	trueVar := true

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Uploaded:          &trueVar,
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateReady, newObj.Status.State)
	require.True(t, newObj.Status.Available)
}
