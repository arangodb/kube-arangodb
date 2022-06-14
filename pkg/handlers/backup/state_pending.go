//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
)

func statePendingHandler(h *handler, backup *backupApi.ArangoBackup) (*backupApi.ArangoBackupStatus, error) {
	_, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return nil, err
	}

	states, err := countBackupStates(backup, h.client.BackupV1().ArangoBackups(backup.Namespace))
	if err != nil {
		return nil, err
	}

	if l := states.get(backupApi.ArangoBackupStateScheduled, backupApi.ArangoBackupStateCreate, backupApi.ArangoBackupStateDownload, backupApi.ArangoBackupStateDownloading); len(l) > 0 {
		return wrapUpdateStatus(backup,
			updateStatusState(backupApi.ArangoBackupStatePending, "backup already in process"))
	}

	return wrapUpdateStatus(backup,
		updateStatusState(backupApi.ArangoBackupStateScheduled, ""))
}
