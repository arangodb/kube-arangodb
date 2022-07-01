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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/uuid"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	deploymentType "github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
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
	checkBackup(t, newObj, backupApi.ArangoBackupStateFailed, false)

	require.Equal(t, newObj.Status.Message,
		createStateMessage(backupApi.ArangoBackupStatePending, backupApi.ArangoBackupStateFailed,
			fmt.Sprintf("%s \"%s\" not found", deploymentType.ArangoDeploymentCRDName, obj.Name)))
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
	checkBackup(t, newObj, backupApi.ArangoBackupStateScheduled, false)
}

func Test_State_Pending_WithUploadRunning(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStatePending)

	uploading := newArangoBackup(deployment.GetName(), deployment.GetNamespace(), string(uuid.NewUUID()), backupApi.ArangoBackupStateUploading)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj, uploading)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateScheduled, false)
}

func Test_State_Pending_WithScheduled(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStatePending)

	uploading := newArangoBackup(deployment.GetName(), deployment.GetNamespace(), string(uuid.NewUUID()), backupApi.ArangoBackupStateScheduled)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj, uploading)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStatePending, false)
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
		checkBackup(t, newObj, backupApi.ArangoBackupStateScheduled, false)
	})

	t.Run("Second backup object", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj2)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj2)
		checkBackup(t, newObj, backupApi.ArangoBackupStatePending, false)
		require.Equal(t, newObj.Status.Message, "backup already in process")
	})
}

func Test_State_Pending_KeepPendingWithForcedRunning(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	name := string(uuid.NewUUID())

	deployment := newArangoDeployment(name, name)
	size := 128
	objects := make([]*backupApi.ArangoBackup, size)
	for id := range objects {
		objects[id] = newArangoBackup(name, name, string(uuid.NewUUID()), backupApi.ArangoBackupStatePending)
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, objects...)

	w := sync.WaitGroup{}
	w.Add(size)
	for _, backup := range objects {
		go func(b *backupApi.ArangoBackup) {
			defer w.Done()
			require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, b)))
		}(backup)
	}

	// Assert
	w.Wait()

	pending := 0
	scheduled := 0

	for _, object := range objects {
		newObj := refreshArangoBackup(t, handler, object)

		switch newObj.Status.State {
		case backupApi.ArangoBackupStatePending:
			pending++
		case backupApi.ArangoBackupStateScheduled:
			scheduled++
		default:
			require.Fail(t, "Unknown state", newObj.Status.State)
		}
	}

	require.Equal(t, 1, scheduled)
	require.Equal(t, size-1, pending)
}
