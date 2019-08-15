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

	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_State_Downloading_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, database.ArangoBackupStateDownloading)
	wrapperConnectionIssues(t, database.ArangoBackupStateDownloading)
}

func Test_State_Downloading_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateDownloading)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	progress, err := mock.Download(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	obj.Spec.Download = &database.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: database.ArangoBackupSpecOperation{
			RepositoryURL: "S3 URL",
		},
		ID: string(backupMeta.ID),
	}

	obj.Status.Progress = &database.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("Restore percent", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateDownloading, newObj.Status.State)
		require.Equal(t, fmt.Sprintf("%d%%", 0), newObj.Status.Progress.Progress)
		require.Equal(t, obj.Status.Progress.JobID, newObj.Status.Progress.JobID)

		require.False(t, newObj.Status.Available)
	})

	t.Run("Restore percent after update", func(t *testing.T) {
		p := 55
		mock.progresses[progress] = ArangoBackupProgress{
			Progress: p,
		}

		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateDownloading, newObj.Status.State)
		require.Equal(t, fmt.Sprintf("%d%%", p), newObj.Status.Progress.Progress)
		require.Equal(t, fmt.Sprintf("%s", progress), newObj.Status.Progress.JobID)

		require.False(t, newObj.Status.Available)
	})

	t.Run("Finished", func(t *testing.T) {
		mock.progresses[progress] = ArangoBackupProgress{
			Completed: true,
		}

		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateReady, newObj.Status.State)
		require.Nil(t, newObj.Status.Progress)

		require.True(t, newObj.Status.Available)
	})
}

func Test_State_Downloading_FailedDownload(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateDownloading)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	progress, err := mock.Download(backupMeta.ID)
	require.NoError(t, err)

	errorMsg := "error"
	mock.progresses[progress] = ArangoBackupProgress{
		Failed:      true,
		FailMessage: errorMsg,
	}

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	obj.Spec.Download = &database.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: database.ArangoBackupSpecOperation{
			RepositoryURL: "S3 URL",
		},
		ID: string(backupMeta.ID),
	}

	obj.Status.Progress = &database.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateFailed, newObj.Status.State)
	require.Equal(t, fmt.Sprintf("download failed with error: %s", errorMsg), newObj.Status.Message)
	require.Nil(t, newObj.Status.Progress)

	require.False(t, newObj.Status.Available)
}

func Test_State_Downloading_FailedProgress(t *testing.T) {
	// Arrange
	errorMsg := "progress error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		progressError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateDownloading)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	progress, err := mock.Download(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	obj.Spec.Download = &database.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: database.ArangoBackupSpecOperation{
			RepositoryURL: "S3 URL",
		},
		ID: string(backupMeta.ID),
	}

	obj.Status.Progress = &database.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateFailed, newObj.Status.State)
	require.Equal(t, errorMsg, newObj.Status.Message)
	require.Nil(t, newObj.Status.Progress)

	require.False(t, newObj.Status.Available)
}

func Test_State_Downloading_TemporaryFailedProgress(t *testing.T) {
	// Arrange
	errorMsg := "progress error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError: true,
		progressError:    errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateDownloading)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	progress, err := mock.Download(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	obj.Spec.Download = &database.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: database.ArangoBackupSpecOperation{
			RepositoryURL: "S3 URL",
		},
		ID: string(backupMeta.ID),
	}

	obj.Status.Progress = &database.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err = handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}
