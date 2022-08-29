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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func Test_State_Unavailable_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateUnavailable)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateUnavailable)
}

func Test_State_Unavailable_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUnavailable)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
	compareBackupMeta(t, backupMeta, newObj)
}

func Test_State_Unavailable_Keep(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUnavailable)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	backupMeta.Available = false

	mock.state.backups[createResponse.ID] = backupMeta

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUnavailable, false)
	compareBackupMeta(t, backupMeta, newObj)
}

func Test_State_Unavailable_Update(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUnavailable)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	backupMeta.Available = false

	mock.state.backups[createResponse.ID] = backupMeta

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("First iteration", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateUnavailable, false)
		compareBackupMeta(t, backupMeta, newObj)
	})

	t.Run("Second iteration", func(t *testing.T) {
		backupMeta.SizeInBytes = 123
		mock.state.backups[backupMeta.ID] = backupMeta

		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateUnavailable, false)
		compareBackupMeta(t, backupMeta, newObj)
		require.Equal(t, uint64(123), newObj.Status.Backup.SizeInBytes)
	})

	t.Run("Do nothing", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateUnavailable, false)
		compareBackupMeta(t, backupMeta, newObj)
	})
}

func Test_State_Unavailable_TemporaryGetFailed(t *testing.T) {
	// Arrange
	error := newTemporaryErrorf("error")
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUnavailable)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUnavailable, false)
}

func Test_State_Unavailable_FatalGetFailed(t *testing.T) {
	// Arrange
	error := newFatalErrorf("error")
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUnavailable)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUnavailable, false)
}

func Test_State_Unavailable_MissingBackup(t *testing.T) {
	// Arrange
	error := driver.ArangoError{
		Code: 404,
	}
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUnavailable)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateDeleted, false)
}
