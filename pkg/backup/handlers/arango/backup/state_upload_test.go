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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	getError    = "get error"
	uploadError = "upload error"
)

func Test_State_Upload_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateUpload)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateUpload)
}

func Test_State_Upload_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateUpload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateUploading)

	require.NotNil(t, newObj.Status.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 1)
	require.Equal(t, progresses[0], newObj.Status.Progress.JobID)

	require.True(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Backup)
	require.Equal(t, string(backupMeta.ID), newObj.Status.Backup.ID)
	require.Equal(t, backupMeta.Version, newObj.Status.Backup.Version)
}

func Test_State_Upload_Failed(t *testing.T) {
	// Arrange
	checks := map[string]mockErrorsArangoClientBackup{
		"get": {
			getError: getError,
		},
		"upload": {
			uploadError: uploadError,
		},
	}

	for name, c := range checks {
		t.Run(name, func(t *testing.T) {
			// Arrange
			handler, mock := newErrorsFakeHandler(c)

			obj, deployment := newObjectSet(backupApi.ArangoBackupStateUpload)

			backupMeta, err := mock.Create()
			require.NoError(t, err)

			obj.Status.Backup = &backupApi.ArangoBackupDetails{
				ID:                string(backupMeta.ID),
				Version:           backupMeta.Version,
				CreationTimestamp: meta.Now(),
			}

			// Act
			createArangoDeployment(t, handler, deployment)
			createArangoBackup(t, handler, obj)

			require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

			// Assert
			newObj := refreshArangoBackup(t, handler, obj)
			require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateFailed)

			require.Nil(t, newObj.Status.Progress)
			progresses := mock.getProgressIDs()
			require.Len(t, progresses, 0)

			require.False(t, newObj.Status.Available)

			require.NotNil(t, newObj.Status.Backup)
			require.Equal(t, string(backupMeta.ID), newObj.Status.Backup.ID)
			require.Equal(t, backupMeta.Version, newObj.Status.Backup.Version)
		})
	}
}

func Test_State_Upload_TemporaryFailed(t *testing.T) {
	// Arrange
	checks := map[string]mockErrorsArangoClientBackup{
		"get": {
			getError: getError,

			isTemporaryError: true,
		},
		"upload": {
			uploadError: uploadError,

			isTemporaryError: true,
		},
	}

	for name, c := range checks {
		t.Run(name, func(t *testing.T) {
			// Arrange
			handler, mock := newErrorsFakeHandler(c)

			obj, deployment := newObjectSet(backupApi.ArangoBackupStateUpload)

			backupMeta, err := mock.Create()
			require.NoError(t, err)

			obj.Status.Backup = &backupApi.ArangoBackupDetails{
				ID:                string(backupMeta.ID),
				Version:           backupMeta.Version,
				CreationTimestamp: meta.Now(),
			}

			// Act
			createArangoDeployment(t, handler, deployment)
			createArangoBackup(t, handler, obj)

			err = handler.Handle(newItemFromBackup(operation.Update, obj))

			// Assert
			require.Error(t, err)
			require.True(t, IsTemporaryError(err))
		})
	}
}
