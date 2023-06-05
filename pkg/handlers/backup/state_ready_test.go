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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

func Test_State_Ready_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateReady)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateReady)
}

func Test_State_Ready_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
	compareBackupMeta(t, backupMeta, newObj)
}

func Test_State_Ready_Unavailable(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	backupMeta.Available = false

	mock.state.backups[createResponse.ID] = backupMeta

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUnavailable, false)
	compareBackupMeta(t, backupMeta, newObj)
}

func Test_State_Ready_ServerDown(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	backupMeta.NumberOfPiecesPresent = backupMeta.NumberOfDBServers - 1

	mock.state.backups[createResponse.ID] = backupMeta

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUnavailable, false)
	compareBackupMeta(t, backupMeta, newObj)
}

func Test_State_Ready_Success_Update(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("First iteration", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
		compareBackupMeta(t, backupMeta, newObj)
	})

	t.Run("Second iteration", func(t *testing.T) {
		backupMeta.SizeInBytes = 123
		mock.state.backups[backupMeta.ID] = backupMeta

		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
		compareBackupMeta(t, backupMeta, newObj)
		require.Equal(t, uint64(123), newObj.Status.Backup.SizeInBytes)
	})

	t.Run("Do nothing", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
		compareBackupMeta(t, backupMeta, newObj)
	})
}

func Test_State_Ready_TemporaryGetFailed(t *testing.T) {
	// Arrange
	error := newTemporaryErrorf("error")
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
}

func Test_State_Ready_FatalGetFailed(t *testing.T) {
	// Arrange
	error := newFatalErrorf("error")
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
}

func Test_State_Ready_MissingBackup(t *testing.T) {
	// Arrange
	error := driver.ArangoError{
		Code: 404,
	}
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: error,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateDeleted, false)
}

func Test_State_Ready_Upload(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)
	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "Any",
	}

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUpload, true)
	compareBackupMeta(t, backupMeta, newObj)
}

func Test_State_Ready_DownloadDoNothing(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)
	obj.Spec.Download = &backupApi.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "any",
		},
		ID: "some",
	}

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, &backupApi.ArangoBackupDetails{
		Downloaded: util.NewType[bool](true),
	})

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
	compareBackupMeta(t, backupMeta, newObj)
	require.NotNil(t, newObj.Status.Backup.Downloaded)
	require.True(t, *newObj.Status.Backup.Downloaded)
}

func Test_State_Ready_DoUploadDownloadedBackup(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)
	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "Any",
	}

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, &backupApi.ArangoBackupDetails{
		Downloaded: util.NewType[bool](true),
		Uploaded:   util.NewType[bool](false),
	})

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUpload, true)
	require.NotNil(t, newObj.Status.Backup.Downloaded)
	require.True(t, *newObj.Status.Backup.Downloaded)
}

func Test_State_Ready_DoNotReUploadBackup(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)
	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "Any",
	}

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, &backupApi.ArangoBackupDetails{
		Uploaded: util.NewType[bool](true),
	})

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
	compareBackupMeta(t, backupMeta, newObj)
	require.NotNil(t, newObj.Status.Backup.Uploaded)
	require.True(t, *newObj.Status.Backup.Uploaded)
}

func Test_State_Ready_RemoveUploadedFlag(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, &backupApi.ArangoBackupDetails{
		Uploaded: util.NewType[bool](true),
	})

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
	require.Nil(t, newObj.Status.Backup.Uploaded)
}

func Test_State_Ready_KeepPendingWithForcedRunning(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	name := string(uuid.NewUUID())

	deployment := newArangoDeployment(name, name)
	size := 128
	objects := make([]*backupApi.ArangoBackup, size)
	for id := range objects {
		createResponse, err := mock.Create()
		require.NoError(t, err)

		backupMeta, err := mock.Get(createResponse.ID)
		require.NoError(t, err)

		obj := newArangoBackup(name, name, string(uuid.NewUUID()), backupApi.ArangoBackupStateReady)

		obj.Status.Backup = createBackupFromMeta(backupMeta, nil)

		obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "s3://test",
		}
		obj.Status.Available = true

		objects[id] = obj
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

	ready := 0
	upload := 0

	for _, object := range objects {
		newObj := refreshArangoBackup(t, handler, object)

		switch newObj.Status.State {
		case backupApi.ArangoBackupStateReady:
			ready++
		case backupApi.ArangoBackupStateUpload:
			upload++
		default:
			require.Fail(t, "Unknown state", newObj.Status.State)
		}
	}

	require.Equal(t, globals.DefaultBackupConcurrentUploads, upload)
	require.Equal(t, size-globals.DefaultBackupConcurrentUploads, ready)
}

func Test_State_Ready_KeepPendingWithForcedRunningSameId(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	name := string(uuid.NewUUID())

	createResponse, err := mock.Create()
	require.NoError(t, err)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	deployment := newArangoDeployment(name, name)
	size := 128
	objects := make([]*backupApi.ArangoBackup, size)
	for id := range objects {

		obj := newArangoBackup(name, name, string(uuid.NewUUID()), backupApi.ArangoBackupStateReady)

		obj.Status.Backup = createBackupFromMeta(backupMeta, nil)
		obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "s3://test",
		}
		obj.Status.Available = true

		objects[id] = obj
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

	ready := 0
	upload := 0

	for _, object := range objects {
		newObj := refreshArangoBackup(t, handler, object)

		switch newObj.Status.State {
		case backupApi.ArangoBackupStateReady:
			ready++
		case backupApi.ArangoBackupStateUpload:
			upload++
		default:
			require.Fail(t, "Unknown state", newObj.Status.State)
		}
	}

	require.Equal(t, 1, upload)
	require.Equal(t, size-1, ready)
}

func Test_State_Ready_Concurrent_Queued(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	createResponse, err := mock.Create()
	require.NoError(t, err)

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)
	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "Any",
	}

	size := globals.DefaultBackupConcurrentUploads
	objects := make([]*backupApi.ArangoBackup, size)
	for id := range objects {
		createResponse, err := mock.Create()
		require.NoError(t, err)

		backupMeta, err := mock.Get(createResponse.ID)
		require.NoError(t, err)

		obj := newArangoBackup(deployment.GetName(), deployment.GetNamespace(), string(uuid.NewUUID()), backupApi.ArangoBackupStateUploading)

		obj.Status.Backup = createBackupFromMeta(backupMeta, nil)
		obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "s3://test",
		}
		obj.Status.Available = true

		objects[id] = obj
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, objects...)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateReady, true)
	compareBackupMeta(t, backupMeta, newObj)
}

func Test_State_Ready_Concurrent_Started(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	createResponse, err := mock.Create()
	require.NoError(t, err)

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	backupMeta, err := mock.Get(createResponse.ID)
	require.NoError(t, err)

	obj.Status.Backup = createBackupFromMeta(backupMeta, nil)
	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "Any",
	}

	size := globals.DefaultBackupConcurrentUploads - 1
	objects := make([]*backupApi.ArangoBackup, size)
	for id := range objects {
		createResponse, err := mock.Create()
		require.NoError(t, err)

		backupMeta, err := mock.Get(createResponse.ID)
		require.NoError(t, err)

		obj := newArangoBackup(deployment.GetName(), deployment.GetNamespace(), string(uuid.NewUUID()), backupApi.ArangoBackupStateUploading)

		obj.Status.Backup = createBackupFromMeta(backupMeta, nil)
		obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
			RepositoryURL: "s3://test",
		}
		obj.Status.Available = true

		objects[id] = obj
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, objects...)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	checkBackup(t, newObj, backupApi.ArangoBackupStateUpload, true)
	compareBackupMeta(t, backupMeta, newObj)
}
