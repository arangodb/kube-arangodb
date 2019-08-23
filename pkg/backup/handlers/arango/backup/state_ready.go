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

	"github.com/arangodb/go-driver"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
)

func stateReadyHandler(h *handler, backup *backupApi.ArangoBackup) (backupApi.ArangoBackupStatus, error) {
	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return createFailedState(err, backup.Status), nil
	}

	client, err := h.arangoClientFactory(deployment, backup)
	if err != nil {
		return backupApi.ArangoBackupStatus{}, NewTemporaryError("unable to create client: %s", err.Error())
	}

	if backup.Status.Backup == nil {
		return createFailedState(fmt.Errorf("backup details are missing"), backup.Status), nil
	}

	_, err = client.Get(driver.BackupID(backup.Status.Backup.ID))
	if err != nil {
		if IsTemporaryError(err) {
			return switchTemporaryError(err, backup.Status)
		}
		// Go into deleted state
		return backupApi.ArangoBackupStatus{
			ArangoBackupState: backupApi.ArangoBackupState{
				State: backupApi.ArangoBackupStateDeleted,
			},
			Backup: backup.Status.Backup,
		}, nil
	}

	// Check if upload flag was specified later in runtime
	if backup.Spec.Upload != nil && backup.Status.Backup.Uploaded == nil {
		return backupApi.ArangoBackupStatus{
			Available:         true,
			ArangoBackupState: newState(backupApi.ArangoBackupStateUpload, "", nil),
			Backup:            backup.Status.Backup,
		}, nil
	}

	// Remove old upload flag
	if backup.Spec.Upload == nil && backup.Status.Backup.Uploaded != nil {
		newBackup := backup.Status.Backup.DeepCopy()
		newBackup.Uploaded = nil
		return backupApi.ArangoBackupStatus{
			Available:         true,
			ArangoBackupState: newState(backupApi.ArangoBackupStateReady, "", nil),
			Backup:            newBackup,
		}, nil
	}

	return backupApi.ArangoBackupStatus{
		Available:         true,
		ArangoBackupState: newState(backupApi.ArangoBackupStateReady, "", nil),
		Backup:            backup.Status.Backup,
	}, nil
}
