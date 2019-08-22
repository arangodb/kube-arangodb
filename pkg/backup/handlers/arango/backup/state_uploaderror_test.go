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
	"time"

	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/stretchr/testify/require"
)

func Test_State_UploadError_Reschedule(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateUploadError)

	obj.Spec.Upload = &database.ArangoBackupSpecOperation{
		RepositoryURL: "S3 URL",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	trueVar := true

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Uploaded:          &trueVar,
	}

	obj.Status.Time.Time = time.Now().Add(-2 * downloadDelay)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, database.ArangoBackupStateUpload)

	require.True(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Backup)
	require.Equal(t, obj.Status.Backup, newObj.Status.Backup)
}

func Test_State_UploadError_Wait(t *testing.T) {
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateUploadError)

	obj.Spec.Upload = &database.ArangoBackupSpecOperation{
		RepositoryURL: "S3 URL",
	}

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	trueVar := true

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Uploaded:          &trueVar,
	}

	obj.Status.Time.Time = time.Now().Add(2 * downloadDelay)
	obj.Status.Message = "message"

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, database.ArangoBackupStateUploadError)

	require.True(t, newObj.Status.Available)
	require.Equal(t, "message", newObj.Status.Message)

	require.NotNil(t, newObj.Status.Backup)
	require.Equal(t, obj.Status.Backup, newObj.Status.Backup)
}

func Test_State_UploadError_BackToReady(t *testing.T) {
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateUploadError)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	trueVar := true

	obj.Status.Backup = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
		Uploaded:          &trueVar,
	}

	obj.Status.Time.Time = time.Now().Add(2 * downloadDelay)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operation.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, database.ArangoBackupStateReady)

	require.True(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Backup)
	require.Equal(t, obj.Status.Backup, newObj.Status.Backup)
}
