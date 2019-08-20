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
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/state"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	notProgressStates = []state.State{
		database.ArangoBackupStatePending,
		database.ArangoBackupStateNone,
		database.ArangoBackupStateFailed,
		database.ArangoBackupStateDeleted,
		database.ArangoBackupStateReady,
	}
)

func inProgress(backup *database.ArangoBackup) bool {
	for _, state := range notProgressStates {
		if state == backup.Status.State {
			return false
		}
	}

	return true
}

func statePendingHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	_, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return createFailedState(err, backup.Status), nil
	}

	// Ensure that only specified number of processes are running
	backups, err := h.client.DatabaseV1alpha().ArangoBackups(backup.Namespace).List(meta.ListOptions{})
	if err != nil {
		return database.ArangoBackupStatus{}, err
	}

	count := 0
	for _, presentBackup := range backups.Items {
		if presentBackup.Name == backup.Name {
			break
		}

		if presentBackup.Spec.Deployment.Name != backup.Spec.Deployment.Name {
			break
		}

		if !inProgress(&presentBackup) {
			continue
		}

		count++
	}

	if count >= 1 {
		return database.ArangoBackupStatus{
			ArangoBackupState: newState(database.ArangoBackupStatePending, "backup already in process", nil),
		}, nil
	}

	return database.ArangoBackupStatus{
		ArangoBackupState: newState(database.ArangoBackupStateScheduled, "", nil),
	}, nil
}
