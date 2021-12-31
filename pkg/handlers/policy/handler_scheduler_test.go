//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

package policy

import (
	"testing"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func Test_Scheduler_Schedule(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())

	policy := newArangoBackupPolicy("* * * */2 *", namespace, name, map[string]string{}, backupApi.ArangoBackupTemplate{})

	database := newArangoDeployment(namespace, map[string]string{
		"test": "me",
	})

	// Act
	createArangoBackupPolicy(t, handler, policy)
	createArangoDeployment(t, handler, database)

	require.NoError(t, handler.Handle(newItemFromBackupPolicy(operation.Update, policy)))

	// Assert
	newPolicy := refreshArangoBackupPolicy(t, handler, policy)
	require.Empty(t, newPolicy.Status.Message)
	require.True(t, newPolicy.Status.Scheduled.Unix() > time.Now().Unix())

	backups := listArangoBackups(t, handler, namespace)
	require.Len(t, backups, 0)
}

func Test_Scheduler_InvalidSchedule(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())

	policy := newArangoBackupPolicy("", namespace, name, map[string]string{}, backupApi.ArangoBackupTemplate{})

	database := newArangoDeployment(namespace, map[string]string{})

	// Act
	createArangoBackupPolicy(t, handler, policy)
	createArangoDeployment(t, handler, database)

	require.NoError(t, handler.Handle(newItemFromBackupPolicy(operation.Update, policy)))

	// Assert
	newPolicy := refreshArangoBackupPolicy(t, handler, policy)
	require.NotNil(t, newPolicy.Status.Message)
	require.Equal(t, "Validation error: error while parsing expr: Empty spec string", newPolicy.Status.Message)

	backups := listArangoBackups(t, handler, namespace)
	require.Len(t, backups, 0)
}

func Test_Scheduler_Valid_OneObject_SelectAll(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())

	policy := newArangoBackupPolicy("* * * */2 *", namespace, name, map[string]string{}, backupApi.ArangoBackupTemplate{})
	policy.Status.Scheduled = meta.Time{
		Time: time.Now().Add(-1 * time.Hour),
	}

	database := newArangoDeployment(namespace, map[string]string{
		"test": "me",
	})

	// Act
	createArangoBackupPolicy(t, handler, policy)
	createArangoDeployment(t, handler, database)

	require.NoError(t, handler.Handle(newItemFromBackupPolicy(operation.Update, policy)))

	// Assert
	newPolicy := refreshArangoBackupPolicy(t, handler, policy)
	require.Empty(t, newPolicy.Status.Message)
	require.True(t, newPolicy.Status.Scheduled.Unix() > time.Now().Unix())

	backups := listArangoBackups(t, handler, namespace)
	require.Len(t, backups, 1)

	isInList(t, backups, database)
	require.NotNil(t, backups[0].Spec.PolicyName)
	require.Equal(t, policy.Name, *backups[0].Spec.PolicyName)
}

func Test_Scheduler_Valid_OneObject_Selector(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())

	selectors := map[string]string{
		"SELECTOR": string(uuid.NewUUID()),
	}

	policy := newArangoBackupPolicy("* * * */2 *", namespace, name, selectors, backupApi.ArangoBackupTemplate{})
	policy.Status.Scheduled = meta.Time{
		Time: time.Now().Add(-1 * time.Hour),
	}

	database := newArangoDeployment(namespace, selectors)
	database2 := newArangoDeployment(namespace, map[string]string{})

	// Act
	createArangoBackupPolicy(t, handler, policy)
	createArangoDeployment(t, handler, database, database2)

	require.NoError(t, handler.Handle(newItemFromBackupPolicy(operation.Update, policy)))

	// Assert
	newPolicy := refreshArangoBackupPolicy(t, handler, policy)
	require.Empty(t, newPolicy.Status.Message)
	require.True(t, newPolicy.Status.Scheduled.Unix() > time.Now().Unix())

	backups := listArangoBackups(t, handler, namespace)
	require.Len(t, backups, 1)

	isInList(t, backups, database)
	require.NotNil(t, backups[0].Spec.PolicyName)
	require.Equal(t, policy.Name, *backups[0].Spec.PolicyName)
}

func Test_Scheduler_Valid_MultipleObject_Selector(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())

	selectors := map[string]string{
		"SELECTOR": string(uuid.NewUUID()),
	}

	policy := newArangoBackupPolicy("* * * */2 *", namespace, name, selectors, backupApi.ArangoBackupTemplate{})
	policy.Status.Scheduled = meta.Time{
		Time: time.Now().Add(-1 * time.Hour),
	}

	database := newArangoDeployment(namespace, selectors)
	database2 := newArangoDeployment(namespace, selectors)

	// Act
	createArangoBackupPolicy(t, handler, policy)
	createArangoDeployment(t, handler, database, database2)

	require.NoError(t, handler.Handle(newItemFromBackupPolicy(operation.Update, policy)))

	// Assert
	newPolicy := refreshArangoBackupPolicy(t, handler, policy)
	require.Empty(t, newPolicy.Status.Message)
	require.True(t, newPolicy.Status.Scheduled.Unix() > time.Now().Unix())

	backups := listArangoBackups(t, handler, namespace)
	require.Len(t, backups, 2)

	isInList(t, backups, database)
	isInList(t, backups, database2)
	require.NotNil(t, backups[0].Spec.PolicyName)
	require.Equal(t, policy.Name, *backups[0].Spec.PolicyName)
	require.NotNil(t, backups[1].Spec.PolicyName)
	require.Equal(t, policy.Name, *backups[1].Spec.PolicyName)
}

func Test_Reschedule(t *testing.T) {
	// Arrange
	handler := newFakeHandler()

	name := string(uuid.NewUUID())
	namespace := string(uuid.NewUUID())

	selectors := map[string]string{
		"SELECTOR": string(uuid.NewUUID()),
	}

	policy := newArangoBackupPolicy("* 13 * * *", namespace, name, selectors, backupApi.ArangoBackupTemplate{})

	// Act
	createArangoBackupPolicy(t, handler, policy)

	t.Run("First schedule", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackupPolicy(operation.Update, policy)))

		// Assert
		newPolicy := refreshArangoBackupPolicy(t, handler, policy)
		require.Empty(t, newPolicy.Status.Message)

		require.Equal(t, 13, newPolicy.Status.Scheduled.Hour())
	})

	t.Run("First schedule - second iteration", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackupPolicy(operation.Update, policy)))

		// Assert
		newPolicy := refreshArangoBackupPolicy(t, handler, policy)
		require.Empty(t, newPolicy.Status.Message)

		require.Equal(t, 13, newPolicy.Status.Scheduled.Hour())
	})

	t.Run("Change schedule", func(t *testing.T) {
		policy = refreshArangoBackupPolicy(t, handler, policy)
		policy.Spec.Schedule = "3 3 * * *"
		updateArangoBackupPolicy(t, handler, policy)

		require.NoError(t, handler.Handle(newItemFromBackupPolicy(operation.Update, policy)))

		// Assert
		newPolicy := refreshArangoBackupPolicy(t, handler, policy)
		require.Empty(t, newPolicy.Status.Message)

		require.Equal(t, 3, newPolicy.Status.Scheduled.Hour())
		require.Equal(t, 3, newPolicy.Status.Scheduled.Minute())
	})
}

func Test_Validate(t *testing.T) {
	acceptedSchedules := []string{
		"0 0 * * MON,TUE,WED,THU,FRI",
		"* * * * *",
	}

	for _, c := range acceptedSchedules {
		t.Run(c, func(t *testing.T) {
			// Arrange
			handler := newFakeHandler()

			name := string(uuid.NewUUID())
			namespace := string(uuid.NewUUID())

			selectors := map[string]string{
				"SELECTOR": string(uuid.NewUUID()),
			}

			policy := newArangoBackupPolicy(c, namespace, name, selectors, backupApi.ArangoBackupTemplate{})

			require.NoError(t, policy.Validate())

			// Act
			createArangoBackupPolicy(t, handler, policy)

			require.NoError(t, handler.Handle(newItemFromBackupPolicy(operation.Update, policy)))

			// Assert
			newPolicy := refreshArangoBackupPolicy(t, handler, policy)
			require.Empty(t, newPolicy.Status.Message)
			require.NotEmpty(t, newPolicy.Status.Scheduled)
		})
	}
}
