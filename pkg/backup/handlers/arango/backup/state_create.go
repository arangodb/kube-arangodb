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
	"github.com/arangodb/go-driver"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func stateCreateHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return createFailedState(err, backup.Status), nil
	}

	client, err := h.arangoClientFactory(deployment, backup)
	if err != nil {
		return database.ArangoBackupStatus{}, NewTemporaryError("unable to create client: %s", err.Error())
	}

	var backupMeta driver.BackupMeta

	// Try to recover old backup. If old backup is missing go into deleted state
	if backup.Status.Details != nil {
		backupMeta, err = client.Get(driver.BackupID(backup.Status.Details.ID))
		if err != nil {
			if IsTemporaryError(err) {
				return switchTemporaryError(err, backup.Status)
			}

			// Go into deleted state
			return database.ArangoBackupStatus{
				ArangoBackupState: database.ArangoBackupState{
					State: database.ArangoBackupStateDeleted,
				},
				Details: backup.Status.Details,
			}, nil
		}
	} else {
		backupMeta, err = client.Create()
		if err != nil {
			return switchTemporaryError(err, backup.Status)
		}
	}

	details := &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: meta.Now(),
	}

	if backup.Spec.Upload != nil {
		return database.ArangoBackupStatus{
			Available: true,
			ArangoBackupState: database.ArangoBackupState{
				State: database.ArangoBackupStateUpload,
			},
			Details: details,
		}, nil
	}

	return database.ArangoBackupStatus{
		Available: true,
		ArangoBackupState: database.ArangoBackupState{
			State: database.ArangoBackupStateReady,
		},
		Details: details,
	}, nil
}
