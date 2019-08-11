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
)

func Test_State_Upload_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, database.ArangoBackupStateUpload)
	wrapperConnectionIssues(t, database.ArangoBackupStateUpload)
}

func Test_State_Upload_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateUpload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateUploading)

	require.NotNil(t, newObj.Status.State.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 1)
	require.Equal(t, progresses[0], newObj.Status.State.Progress.JobID)

	require.True(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Details)
	require.Equal(t, string(backupMeta.ID), newObj.Status.Details.ID)
	require.Equal(t, backupMeta.Version, newObj.Status.Details.Version)
}

func Test_State_Upload_GetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateUpload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateFailed)

	require.Nil(t, newObj.Status.State.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 0)

	require.False(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Details)
	require.Equal(t, string(backupMeta.ID), newObj.Status.Details.ID)
	require.Equal(t, backupMeta.Version, newObj.Status.Details.Version)
}

func Test_State_Upload_TemporaryGetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError: true,

		getError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateUpload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err = handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}

func Test_State_Upload_UploadFailed(t *testing.T) {
	// Arrange
	errorMsg := "upload error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		uploadError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateUpload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateFailed)

	require.Nil(t, newObj.Status.State.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 0)

	require.False(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Details)
	require.Equal(t, string(backupMeta.ID), newObj.Status.Details.ID)
	require.Equal(t, backupMeta.Version, newObj.Status.Details.Version)
}

func Test_State_Upload_TemporaryUploadFailed(t *testing.T) {
	// Arrange
	errorMsg := "upload error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError: true,
		uploadError:      errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateUpload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err = handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}
