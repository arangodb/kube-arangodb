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

func Test_Flow_SuccessHappyPath(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateNone)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("Change from None to Pending", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStatePending, newObj.Status.State)
	})

	t.Run("Change from Pending to Scheduled", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateScheduled, newObj.Status.State)
	})

	t.Run("Change from Scheduled to Create", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateCreate, newObj.Status.State)
	})

	t.Run("Change from Create to Ready", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateReady, newObj.Status.State)
	})

	t.Run("Ensure Ready State Keeps", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateReady, newObj.Status.State)
	})

	t.Run("Change from Ready to Deleted", func(t *testing.T) {
		// Arrange
		mock.errors.getError = "error"

		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateDeleted, newObj.Status.State)
	})
}
