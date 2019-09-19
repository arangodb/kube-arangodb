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

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	"github.com/stretchr/testify/require"
)

func Test_State_Create_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateCreate)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateCreate)
}

func Test_State_Create_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateReady)

	require.NotNil(t, newObj.Status.Backup)

	backups := mock.getIDs()
	require.Len(t, backups, 1)

	require.Equal(t, newObj.Status.Backup.ID, backups[0])
	require.Equal(t, newObj.Status.Backup.Version, mockVersion)

	require.Nil(t, newObj.Status.Backup.PotentiallyInconsistent)
}

func Test_State_Create_SuccessForced(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})
	mock.errors.createForced = true

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateReady)

	require.NotNil(t, newObj.Status.Backup)

	backups := mock.getIDs()
	require.Len(t, backups, 1)

	require.Equal(t, newObj.Status.Backup.ID, backups[0])
	require.Equal(t, newObj.Status.Backup.Version, mockVersion)

	require.NotNil(t, newObj.Status.Backup.PotentiallyInconsistent)
	value := *newObj.Status.Backup.PotentiallyInconsistent
	require.True(t, value)
}

func Test_State_Create_Upload(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)
	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "test",
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateReady)

	require.NotNil(t, newObj.Status.Backup)

	backups := mock.getIDs()
	require.Len(t, backups, 1)

	require.Equal(t, newObj.Status.Backup.ID, backups[0])
	require.Equal(t, newObj.Status.Backup.Version, mockVersion)

	require.True(t, newObj.Status.Available)
}

func Test_State_Create_CreateFailed(t *testing.T) {
	// Arrange
	errorMsg := "create error"
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		createError: errorMsg,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateFailed)
	require.Equal(t, newObj.Status.Message, createFailMessage(backupApi.ArangoBackupStateCreate, errorMsg))

	require.Nil(t, newObj.Status.Backup)

	require.False(t, newObj.Status.Available)
}

func Test_State_Create_TemporaryCreateFailed(t *testing.T) {
	// Arrange
	errorMsg := "create error"
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError: true,
		createError:      errorMsg,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err := handler.Handle(newItemFromBackup(operation.Update, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}
