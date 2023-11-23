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
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_State_Creating_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateCreating)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateCreating)
}

func Test_State_Creating_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(t, backupApi.ArangoBackupStateCreating)

	obj.Status.Progress = &backupApi.ArangoBackupProgress{
		JobID: "jobID",
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("Create in progress, then done", func(t *testing.T) {
		require.NoError(t, handler.Handle(context.Background(), tests.NewItem(t, operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateCreating, false)

		require.NotNil(t, newObj.Status.Progress)

		require.Equal(t, fmt.Sprintf("%d%%", 50), newObj.Status.Progress.Progress)
		require.Equal(t, obj.Status.Progress.JobID, newObj.Status.Progress.JobID)

		mock.state.createDone = true
		require.NoError(t, handler.Handle(context.Background(), tests.NewItem(t, operation.Update, obj)))

		// Assert
		newObj = refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
		require.Nil(t, newObj.Status.Progress)

		backups := mock.getIDs()
		require.Len(t, backups, 1)

		backupMeta, err := mock.Get(driver.BackupID(backups[0]))
		require.NoError(t, err)

		compareBackupMeta(t, backupMeta, newObj)

	})
}

func Test_State_Creating_Failed(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		createError: driver.ArangoError{
			Code: 400,
		},
	})

	obj, deployment := newObjectSet(t, backupApi.ArangoBackupStateCreating)

	obj.Status.Progress = &backupApi.ArangoBackupProgress{
		JobID: "jobID",
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("Create Backup returns error", func(t *testing.T) {
		require.NoError(t, handler.Handle(context.Background(), tests.NewItem(t, operation.Update, obj)))

		// Create error state should be set
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateCreateError, false)
		require.Nil(t, newObj.Status.Progress)

		require.NoError(t, handler.Handle(context.Background(), tests.NewItem(t, operation.Update, obj)))

		// No retry - state should change to failed
		newObj = refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateFailed, false)
		require.Nil(t, newObj.Status.Progress)

	})
}
