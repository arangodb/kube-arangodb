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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func Test_State_Uploading_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateUploading)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateUploading)
	wrapperProgressMissing(t, backupApi.ArangoBackupStateUploading)
}

func Test_State_Uploading_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploading)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "S3 URL",
	}

	obj.Status.Progress = &backupApi.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("Restore percent", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateUploading, true)
		require.Equal(t, fmt.Sprintf("%d%%", 0), newObj.Status.Progress.Progress)
		require.Equal(t, obj.Status.Progress.JobID, newObj.Status.Progress.JobID)
	})

	t.Run("Restore percent after update", func(t *testing.T) {
		p := 55
		mock.state.progresses[progress] = ArangoBackupProgress{
			Progress: p,
		}

		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateUploading, true)
		require.Equal(t, fmt.Sprintf("%d%%", p), newObj.Status.Progress.Progress)
		require.Equal(t, string(progress), newObj.Status.Progress.JobID)
	})

	t.Run("Finished", func(t *testing.T) {
		mock.state.progresses[progress] = ArangoBackupProgress{
			Completed: true,
		}

		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
		require.Nil(t, newObj.Status.Progress)

		require.NotNil(t, newObj.Status.Backup.Uploaded)
		require.True(t, *newObj.Status.Backup.Uploaded)
	})
}

func Test_State_Uploading_FailedUpload(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploading)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	errorMsg := errorString
	mock.state.progresses[progress] = ArangoBackupProgress{
		Failed:      true,
		FailMessage: errorMsg,
	}

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	obj.Status.Progress = &backupApi.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUploadError, true)
	require.Equal(t, fmt.Sprintf("Upload failed with error: %s", errorMsg), newObj.Status.Message)
	require.Nil(t, newObj.Status.Progress)
}

func Test_StateUploading_FailedProgress(t *testing.T) {
	// Arrange
	error := newFatalErrorf("error")
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		progressError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploading)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "S3 URL",
	}

	obj.Status.Progress = &backupApi.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.EqualError(t, handler.Handle(newItemFromBackup(operation.Update, obj)), error.Error())

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, obj.Status, newObj.Status)
}

func Test_State_Uploading_TemporaryFailedProgress(t *testing.T) {
	// Arrange
	error := newTemporaryErrorf("error")
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		progressError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploading)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "S3 URL",
	}

	obj.Status.Progress = &backupApi.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.EqualError(t, handler.Handle(newItemFromBackup(operation.Update, obj)), error.Error())

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, obj.Status, newObj.Status)
}

func Test_State_Uploading_NotFoundProgress(t *testing.T) {
	// Arrange
	error := driver.ArangoError{
		Code: 404,
	}
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		progressError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploading)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "S3 URL",
	}

	obj.Status.Progress = &backupApi.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUploadError, true)
	require.Equal(t, fmt.Sprintf("job with id %s does not exist anymore", progress), newObj.Status.Message)
	require.Nil(t, newObj.Status.Progress)
}

func Test_StateUploading_Abort_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploading)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	obj.Status.Progress = &backupApi.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
}

func Test_StateUploading_Abort_Fail(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		abortError: newTemporaryErrorf("error"),
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUploading)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	obj.Status.Progress = &backupApi.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUploading, true)
}
