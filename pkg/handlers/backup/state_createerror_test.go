//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_State_CreateError_Retry_WhenBackoffEnabled(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(t, backupApi.ArangoBackupStateCreateError)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	obj.Spec.Backoff = &backupApi.ArangoBackupSpecBackOff{
		MaxIterations: util.NewType(1),
	}

	obj.Status.Time.Time = time.Now().Add(-2 * downloadDelay)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(context.Background(), tests.NewItem(t, operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateCreate)
	require.False(t, newObj.Status.Available)
	require.NotNil(t, newObj.Status.Backup)
	require.Equal(t, obj.Status.Backup, newObj.Status.Backup)
}

func Test_State_CreateError_Retry_WhenBackoffDisabled_C1(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(t, backupApi.ArangoBackupStateCreateError)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	obj.Spec.Backoff = &backupApi.ArangoBackupSpecBackOff{}

	obj.Status.Time.Time = time.Now().Add(-2 * downloadDelay)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(context.Background(), tests.NewItem(t, operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateFailed)
	require.Equal(t, newObj.Status.Message, "retries are disabled")
}

func Test_State_CreateError_Retry_WhenBackoffDisabled_C2(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(t, backupApi.ArangoBackupStateCreateError)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	obj.Status.Time.Time = time.Now().Add(-2 * downloadDelay)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(context.Background(), tests.NewItem(t, operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateFailed)
	require.Equal(t, newObj.Status.Message, "retries are disabled")
}

func Test_State_CreateError_Transfer_To_Failed(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(t, backupApi.ArangoBackupStateCreateError)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Backup = &backupApi.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}
	obj.Status.Backoff = &backupApi.ArangoBackupStatusBackOff{
		Iterations: 2,
	}

	obj.Spec.Backoff = &backupApi.ArangoBackupSpecBackOff{
		Iterations:    util.NewType[int](1),
		MaxIterations: util.NewType[int](2),
	}

	obj.Status.Time.Time = time.Now().Add(-2 * downloadDelay)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(context.Background(), tests.NewItem(t, operation.Update, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State, backupApi.ArangoBackupStateFailed)
	require.Equal(t, newObj.Status.Message, "out of Create retries")
}
