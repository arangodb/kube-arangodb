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
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

func stateDownloadHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return createFailedState(err, backup.Status), nil
	}

	client, err := h.arangoClientFactory(deployment, backup)
	if err != nil {
		return database.ArangoBackupStatus{}, NewTemporaryError("unable to create client: %s", err.Error())
	}

	if backup.Status.Details == nil {
		return createFailedState(fmt.Errorf("backup details are missing"), backup.Status), nil
	}

	meta, err := client.Get(driver.BackupID(backup.Status.Details.ID))
	if err != nil {
		return switchTemporaryError(err, backup.Status)
	}

	jobID, err := client.Download(meta.ID)
	if err != nil {
		return switchTemporaryError(err, backup.Status)
	}

	return database.ArangoBackupStatus{
		Available: false,
		ArangoBackupState: database.ArangoBackupState{
			State: database.ArangoBackupStateDownloading,
			Progress: &database.ArangoBackupProgress{
				JobID:    string(jobID),
				Progress: "0%",
			},
		},
		Details: backup.Status.Details,
	}, nil
}
