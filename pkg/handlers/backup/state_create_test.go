//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_State_Create_Common(t *testing.T) {
	*features.AsyncBackupCreation().EnabledPointer() = false
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateCreate)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateCreate)
}

func Test_State_Create_Common_Async(t *testing.T) {
	*features.AsyncBackupCreation().EnabledPointer() = true
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateCreate)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateCreate)
}

func Test_State_Create_Success(t *testing.T) {
	*features.AsyncBackupCreation().EnabledPointer() = false
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)

	backups := mock.getIDs()
	require.Len(t, backups, 1)

	backupMeta, err := mock.Get(driver.BackupID(backups[0]))
	require.NoError(t, err)

	compareBackupMeta(t, backupMeta, newObj)
}

func Test_State_Create_Success_Async(t *testing.T) {
	*features.AsyncBackupCreation().EnabledPointer() = true
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateCreating, false)
}

func Test_State_Create_SuccessForced(t *testing.T) {
	*features.AsyncBackupCreation().EnabledPointer() = false

	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)
	obj.Spec.Options = &backupApi.ArangoBackupSpecOptions{
		AllowInconsistent: util.NewType[bool](true),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)

	backups := mock.getIDs()
	require.Len(t, backups, 1)

	backupMeta, err := mock.Get(driver.BackupID(backups[0]))
	require.NoError(t, err)

	compareBackupMeta(t, backupMeta, newObj)
	require.NotNil(t, newObj.Status.Backup.PotentiallyInconsistent)
	require.True(t, *newObj.Status.Backup.PotentiallyInconsistent)
}

func Test_State_Create_Upload(t *testing.T) {
	*features.AsyncBackupCreation().EnabledPointer() = false

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
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)

	backups := mock.getIDs()
	require.Len(t, backups, 1)

	backupMeta, err := mock.Get(driver.BackupID(backups[0]))
	require.NoError(t, err)

	compareBackupMeta(t, backupMeta, newObj)
}

func Test_State_Create_CreateError(t *testing.T) {
	*features.AsyncBackupCreation().EnabledPointer() = false

	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		createError: newFatalErrorf("error"),
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateCreateError)
	require.Nil(t, newObj.Status.Backup)
	require.False(t, newObj.Status.Available)
}

func Test_State_Create_CreateError_Async(t *testing.T) {
	*features.AsyncBackupCreation().EnabledPointer() = true

	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		createError: newFatalErrorf("error"),
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateCreateError)
	require.Nil(t, newObj.Status.Backup)
	require.False(t, newObj.Status.Available)
}
