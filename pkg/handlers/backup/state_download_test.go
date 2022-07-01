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

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func Test_State_Download_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateDownload)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateDownload)
}

func Test_State_Download_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateDownload)

	obj.Spec.Download = &backupApi.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "S3 URL",
		},
		ID: "test",
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateDownloading, false)

	require.NotNil(t, newObj.Status.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 1)
	require.Equal(t, progresses[0], newObj.Status.Progress.JobID)

	require.Nil(t, newObj.Status.Backup)
}

// Check version
func Test_State_Download_DownloadFailed(t *testing.T) {
	// Arrange
	error := newFatalErrorf("error")
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		downloadError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateDownload)

	obj.Spec.Download = &backupApi.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "S3 URL",
		},
		ID: "test",
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateDownloadError, false)

	require.Nil(t, newObj.Status.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 0)

	require.Nil(t, newObj.Status.Backup)
}

func Test_State_Download_TemporaryDownloadFailed(t *testing.T) {
	// Arrange
	error := newTemporaryErrorf("error")
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		downloadError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateDownload)

	obj.Spec.Download = &backupApi.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "S3 URL",
		},
		ID: "test",
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateDownloadError, false)

	require.Nil(t, newObj.Status.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 0)

	require.Nil(t, newObj.Status.Backup)
}
