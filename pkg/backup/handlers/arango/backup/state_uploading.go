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

func stateUploadingHandler(h *handler, backup *backupApi.ArangoBackup) (backupApi.ArangoBackupStatus, error) {
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

	if backup.Status.Progress == nil {
		return createFailedState(fmt.Errorf("backup progress details are missing"), backup.Status), nil
	}

	details, err := client.Progress(driver.BackupTransferJobID(backup.Status.Progress.JobID))
	if err != nil {
		return switchTemporaryError(err, backup.Status)
	}

	if details.Failed {
		return backupApi.ArangoBackupStatus{
			Available: true,
			ArangoBackupState: newState(backupApi.ArangoBackupStateUploadError,
				fmt.Sprintf("Upload failed with error: %s", details.FailMessage), nil),
			Backup: backup.Status.Backup,
		}, nil
	}

	if details.Completed {
		newDetails := backup.Status.Backup.DeepCopy()

		trueVar := true

		newDetails.Uploaded = &trueVar

		return backupApi.ArangoBackupStatus{
			Available:         true,
			ArangoBackupState: newState(backupApi.ArangoBackupStateReady, "", nil),
			Backup:            newDetails,
		}, nil
	}

	return backupApi.ArangoBackupStatus{
		Available: true,
		ArangoBackupState: newState(backupApi.ArangoBackupStateUploading, "",
			&backupApi.ArangoBackupProgress{
				JobID:    backup.Status.Progress.JobID,
				Progress: fmt.Sprintf("%d%%", details.Progress),
			}),
		Backup: backup.Status.Backup,
	}, nil
}
