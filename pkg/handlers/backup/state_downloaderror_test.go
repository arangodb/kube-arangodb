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
	"time"

	"github.com/stretchr/testify/require"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func Test_State_DownloadError_Reschedule(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateDownloadError)

	obj.Spec.Download = &backupApi.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "S3 URL",
		},
		ID: "test",
	}

	obj.Status.Time.Time = time.Now().Add(-2 * downloadDelay)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStatePending, false)

	require.Nil(t, newObj.Status.Backup)
}

func Test_State_DownloadError_Wait(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateDownloadError)

	obj.Spec.Download = &backupApi.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "S3 URL",
		},
		ID: "test",
	}

	obj.Status.Time.Time = time.Now().Add(2 * downloadDelay)
	obj.Status.Message = "message"

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateDownloadError, false)

	require.Equal(t, "message", newObj.Status.Message)

	require.Nil(t, newObj.Status.Backup)
}
