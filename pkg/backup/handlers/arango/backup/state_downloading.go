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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/go-driver"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

func stateDownloadingHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return createFailedState(err, backup.Status), nil
	}

	client, err := h.arangoClientFactory(deployment, backup)
	if err != nil {
		return database.ArangoBackupStatus{}, NewTemporaryError("unable to create client: %s", err.Error())
	}

	if backup.Status.Progress == nil {
		return createFailedState(fmt.Errorf("backup progress details are missing"), backup.Status), nil
	}

	if backup.Spec.Download == nil {
		return createFailedState(fmt.Errorf("missing field .spec.download"), backup.Status), nil
	}

	if backup.Spec.Download.ID == "" {
		return createFailedState(fmt.Errorf("missing field .spec.download.id"), backup.Status), nil
	}

	details, err := client.Progress(driver.BackupTransferJobID(backup.Status.Progress.JobID))
	if err != nil {
		return switchTemporaryError(err, backup.Status)
	}

	if details.Failed {
		return database.ArangoBackupStatus{
			Available: false,
			ArangoBackupState: newState(database.ArangoBackupStateDownloadError,
				fmt.Sprintf("Download failed with error: %s", details.FailMessage), nil),
		}, nil
	}

	if details.Completed {
		backupMeta, err := client.Get(driver.BackupID(backup.Spec.Download.ID))
		if err != nil {
			return switchTemporaryError(err, backup.Status)
		}

		trueVar := true

		return database.ArangoBackupStatus{
			Available:         true,
			ArangoBackupState: newState(database.ArangoBackupStateReady, "", nil),
			Backup: &database.ArangoBackupDetails{
				ID:                string(backupMeta.ID),
				Version:           backupMeta.Version,
				CreationTimestamp: meta.Now(),
				Downloaded:        &trueVar,
			},
		}, nil
	}

	return database.ArangoBackupStatus{
		Available: false,
		ArangoBackupState: newState(database.ArangoBackupStateDownloading, "",
			&database.ArangoBackupProgress{
				JobID:    backup.Status.Progress.JobID,
				Progress: fmt.Sprintf("%d%%", details.Progress),
			}),
	}, nil
}
