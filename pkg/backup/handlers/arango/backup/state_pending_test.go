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

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/stretchr/testify/require"
)

func Test_State_Pending_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStatePending)
}

func Test_State_Pending_CheckNamespaceIsolation(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStatePending)
	deployment.Namespace = "non-existent"

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateFailed)

	require.Equal(t, newObj.Status.Message, createFailMessage(backupApi.ArangoBackupStatePending, fmt.Sprintf("%s \"%s\" not found", database.ArangoDeploymentCRDName, obj.Name)))
}

func Test_State_Pending_OneBackupObject(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStatePending)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateScheduled)

	require.False(t, newObj.Status.Available)
}

func Test_State_Pending_MultipleBackupObjectWithLimitation(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStatePending)
	obj2, _ := newObjectSet(backupApi.ArangoBackupStatePending)
	obj2.Namespace = obj.Namespace
	obj2.Spec.Deployment.Name = deployment.Name

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj, obj2)

	t.Run("First backup object", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateScheduled)

		require.False(t, newObj.Status.Available)
	})

	t.Run("Second backup object", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj2)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj2)
		require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStatePending)
		require.Equal(t, newObj.Status.Message, "backup already in process")
	})
}
