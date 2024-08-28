//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_ObjectNotFound(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	i := tests.NewItem(t, operation.Add, tests.NewMetaObject[*backupApi.ArangoBackup](t, "none", "none"))

	actions := map[operation.Operation]bool{
		operation.Add:    false,
		operation.Update: false,
		operation.Delete: false,
	}

	// Act
	for operation, shouldFail := range actions {
		t.Run(string(operation), func(t *testing.T) {
			err := handler.Handle(context.Background(), i)

			// Assert
			if shouldFail {
				require.Error(t, err)
				require.True(t, apiErrors.IsNotFound(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func resetFeature(f features.Feature) func() {
	enabled := f.Enabled()

	return func() {
		*f.EnabledPointer() = enabled
	}
}

func Test_Refresh_Cleanup(t *testing.T) {
	defer resetFeature(features.BackupCleanup())()

	// Arrange
	handler, client := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	id := driver.BackupID(uuid.NewUUID())
	client.state.backups = map[driver.BackupID]driver.BackupMeta{
		id: {
			ID:                      id,
			Version:                 "3.12.0",
			DateTime:                time.Now().Add(-time.Hour),
			NumberOfFiles:           123,
			NumberOfDBServers:       3,
			SizeInBytes:             123,
			PotentiallyInconsistent: false,
			Available:               true,
			NumberOfPiecesPresent:   123,
		},
	}

	arangoDeployment := tests.NewMetaObject[*database.ArangoDeployment](t, tests.FakeNamespace, "deployment")

	t.Run("Discover", func(t *testing.T) {
		require.NoError(t, handler.refreshDeployment(arangoDeployment))

		backups, err := handler.client.BackupV1().ArangoBackups(tests.FakeNamespace).List(context.Background(), meta.ListOptions{})
		require.NoError(t, err)
		require.Len(t, backups.Items, 1)
		require.NotNil(t, backups.Items[0].Status.Backup)
		require.EqualValues(t, id, backups.Items[0].Status.Backup.ID)
	})

	t.Run("Without Cleanup Feature", func(t *testing.T) {
		*features.BackupCleanup().EnabledPointer() = false

		require.NoError(t, handler.refreshDeployment(arangoDeployment))

		backups, err := handler.client.BackupV1().ArangoBackups(tests.FakeNamespace).List(context.Background(), meta.ListOptions{})
		require.NoError(t, err)
		require.Len(t, backups.Items, 1)
		require.NotNil(t, backups.Items[0].Status.Backup)
		require.EqualValues(t, id, backups.Items[0].Status.Backup.ID)
	})

	t.Run("With Cleanup Feature", func(t *testing.T) {
		*features.BackupCleanup().EnabledPointer() = true

		require.NoError(t, handler.refreshDeployment(arangoDeployment))

		backups, err := handler.client.BackupV1().ArangoBackups(tests.FakeNamespace).List(context.Background(), meta.ListOptions{})
		require.NoError(t, err)
		require.Len(t, backups.Items, 0)
	})

	t.Run("Do not refresh if backup is creating", func(t *testing.T) {
		// Arrange
		fakeId := driver.BackupID(uuid.NewUUID())
		createBackup := backupApi.ArangoBackup{

			ObjectMeta: meta.ObjectMeta{
				Name: "backup",
			},
			Status: backupApi.ArangoBackupStatus{
				ArangoBackupState: backupApi.ArangoBackupState{
					State: backupApi.ArangoBackupStateCreating,
				},
				Backup: &backupApi.ArangoBackupDetails{
					ID: string(fakeId),
				},
			},
		}
		b, err := handler.client.BackupV1().ArangoBackups(tests.FakeNamespace).Create(context.Background(), &createBackup, meta.CreateOptions{})
		require.NoError(t, err)
		require.NotNil(t, b)
		require.Equal(t, backupApi.ArangoBackupStateCreating, b.Status.State)

		t.Run("Refresh should not happen if there is Backup in creation state", func(t *testing.T) {
			require.NoError(t, handler.refreshDeployment(arangoDeployment))

			backups, err := handler.client.BackupV1().ArangoBackups(tests.FakeNamespace).List(context.Background(), meta.ListOptions{})
			require.NoError(t, err)
			require.Len(t, backups.Items, 1)
			require.NotNil(t, backups.Items[0].Status.Backup)
			require.EqualValues(t, fakeId, backups.Items[0].Status.Backup.ID)
		})

		createBackup.Status.State = backupApi.ArangoBackupStateReady
		b, err = handler.client.BackupV1().ArangoBackups(tests.FakeNamespace).UpdateStatus(context.Background(), &createBackup, meta.UpdateOptions{})
		require.NoError(t, err)
		require.NotNil(t, b)
		require.Equal(t, backupApi.ArangoBackupStateReady, b.Status.State)

		t.Run("Refresh should happen if there is Backup in ready state", func(t *testing.T) {
			require.NoError(t, handler.refreshDeployment(arangoDeployment))

			backups, err := handler.client.BackupV1().ArangoBackups(tests.FakeNamespace).List(context.Background(), meta.ListOptions{})
			require.NoError(t, err)
			require.Len(t, backups.Items, 2)
		})
	})
}
