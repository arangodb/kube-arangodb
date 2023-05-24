//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"fmt"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func stateDownloadingHandler(h *handler, backup *backupApi.ArangoBackup) (*backupApi.ArangoBackupStatus, error) {
	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return nil, err
	}

	client, err := h.arangoClientFactory(deployment, backup)
	if err != nil {
		return nil, newTemporaryError(err)
	}

	if backup.Status.Progress == nil {
		return nil, newFatalErrorf("backup progress details are missing")
	}

	if backup.Spec.Download == nil {
		return nil, newFatalErrorf("missing field .spec.download")
	}

	if backup.Spec.Download.ID == "" {
		return nil, newFatalErrorf("missing field .spec.download.id")
	}

	details, err := client.Progress(driver.BackupTransferJobID(backup.Status.Progress.JobID))
	if err != nil {
		if driver.IsNotFound(err) {
			return wrapUpdateStatus(backup,
				updateStatusState(backupApi.ArangoBackupStateDownloadError,
					"job with id %s does not exist anymore", backup.Status.Progress.JobID),
				cleanStatusJob(),
			)
		}

		return nil, newTemporaryError(err)
	}

	if details.Failed {
		return wrapUpdateStatus(backup,
			updateStatusState(backupApi.ArangoBackupStateDownloadError,
				"Download failed with error: %s", details.FailMessage),
			cleanStatusJob(),
		)
	}

	if details.Completed {
		backupMeta, err := client.Get(driver.BackupID(backup.Spec.Download.ID))
		if err != nil {
			if driver.IsNotFound(err) {
				return wrapUpdateStatus(backup,
					updateStatusState(backupApi.ArangoBackupStateDownloadError,
						"backup is not present after download"),
					cleanStatusJob(),
				)
			}

			return nil, newTemporaryError(err)
		}

		return wrapUpdateStatus(backup,
			updateStatusState(backupApi.ArangoBackupStateReady, ""),
			updateStatusAvailable(true),
			updateStatusBackup(backupMeta),
			updateStatusBackupDownload(util.NewType[bool](true)),
			cleanStatusJob(),
		)
	}

	return wrapUpdateStatus(backup,
		updateStatusState(backupApi.ArangoBackupStateDownloading, ""),
		updateStatusJob(backup.Status.Progress.JobID, fmt.Sprintf("%d%%", details.Progress)),
	)
}
