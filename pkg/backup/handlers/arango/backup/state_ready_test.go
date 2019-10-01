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
	"github.com/arangodb/kube-arangodb/pkg/util"
	"sync"
	"testing"

	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_State_Ready_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, backupApi.ArangoBackupStateReady)
	wrapperConnectionIssues(t, backupApi.ArangoBackupStateReady)
}

func Test_State_Ready_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, obj.Status, newObj.Status)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_GetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: errorMsg,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, backupApi.ArangoBackupStateReady, newObj.Status.State)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_TemporaryGetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError: true,
		getError:         errorMsg,
	})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err = handler.Handle(newItemFromBackup(operation.Update, obj))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, backupApi.ArangoBackupStateReady, newObj.Status.State)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_Upload(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)
	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "Any",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, backupApi.ArangoBackupStateUpload, newObj.Status.State)
	require.True(t, newObj.Status.Available)
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

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, backupApi.ArangoBackupStateReady, newObj.Status.State)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_DoUploadDownloadedBackup(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)
	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "Any",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Downloaded:        util.NewBool(true),
		Uploaded:          util.NewBool(false),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, backupApi.ArangoBackupStateUpload, newObj.Status.State)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_DoNotReUploadBackup(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)
	obj.Spec.Upload = &backupApi.ArangoBackupSpecOperation{
		RepositoryURL: "Any",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Uploaded:          util.NewBool(true),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, backupApi.ArangoBackupStateReady, newObj.Status.State)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_RemoveUploadedFlag(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(backupApi.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Uploaded:          util.NewBool(true),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, backupApi.ArangoBackupStateReady, newObj.Status.State)
	require.True(t, newObj.Status.Available)
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
		backupMeta, err := mock.Create()
		require.NoError(t, err)

		obj := newArangoBackup(name, name, string(uuid.NewUUID()), backupApi.ArangoBackupStateReady)

		obj.Status.Backup = &backupApi.ArangoBackupDetails{
			ID:                string(backupMeta.ID),
			Version:           backupMeta.Version,
			CreationTimestamp: meta.Now(),
		}
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

	require.Equal(t, size, upload)
	require.Equal(t, 0, ready)
}

func Test_State_Ready_KeepPendingWithForcedRunningSameId(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	name := string(uuid.NewUUID())

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	deployment := newArangoDeployment(name, name)
	size := 128
	objects := make([]*backupApi.ArangoBackup, size)
	for id := range objects {

		obj := newArangoBackup(name, name, string(uuid.NewUUID()), backupApi.ArangoBackupStateReady)

		obj.Status.Backup = &backupApi.ArangoBackupDetails{
			ID:                string(backupMeta.ID),
			Version:           backupMeta.Version,
			CreationTimestamp: meta.Now(),
		}
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
