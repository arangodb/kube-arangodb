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

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"
	"github.com/stretchr/testify/require"
)

func Test_Flow_SuccessHappyPath(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateNone)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("Change from None to Pending", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, backupApi.ArangoBackupStatePending, newObj.Status.State)
	})

	t.Run("Change from Pending to Scheduled", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, backupApi.ArangoBackupStateScheduled, newObj.Status.State)
	})

	t.Run("Change from Scheduled to Create", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, backupApi.ArangoBackupStateCreate, newObj.Status.State)
	})

	t.Run("Change from Create to Ready", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, backupApi.ArangoBackupStateReady, newObj.Status.State)
	})

	t.Run("Ensure Ready State Keeps", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, backupApi.ArangoBackupStateReady, newObj.Status.State)
	})

	t.Run("Change from Ready to Deleted", func(t *testing.T) {
		// Arrange
		mock.errors.getError = errorString

		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, backupApi.ArangoBackupStateDeleted, newObj.Status.State)
	})
}
