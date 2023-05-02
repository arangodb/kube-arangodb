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
	"time"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
)

func stateUploadErrorHandler(h *handler, backup *backupApi.ArangoBackup) (*backupApi.ArangoBackupStatus, error) {
	// no more retries - move to failed state
	if !backup.Status.Backoff.ShouldBackoff(backup.Spec.Backoff) {
		return wrapUpdateStatus(backup,
			updateStatusState(backupApi.ArangoBackupStateFailed, "out of Upload retries"),
			cleanStatusJob())
	}

	// if we should retry - move to ready state
	if backup.Spec.Upload == nil ||
		(backup.Status.Backoff.ShouldBackoff(backup.Spec.Backoff) && !backup.Status.Backoff.GetNext().After(time.Now())) {
		return wrapUpdateStatus(backup,
			updateStatusState(backupApi.ArangoBackupStateReady, ""),
			cleanStatusJob(),
			updateStatusAvailable(true))
	}

	// no ready to retry - wait (do not change state)
	return wrapUpdateStatus(backup,
		updateStatusAvailable(true))
}
