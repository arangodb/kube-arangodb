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
	"time"

	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

const (
	uploadDelay = time.Minute
)

func stateUploadErrorHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	// After upload removal go into Ready status
	if backup.Spec.Upload == nil {
		return database.ArangoBackupStatus{
			Available:         true,
			ArangoBackupState: newState(database.ArangoBackupStateReady, "", nil),
			Backup:            backup.Status.Backup.DeepCopy(),
		}, nil
	}

	// Start again upload
	if backup.Status.Time.Time.Add(uploadDelay).Before(time.Now()) {
		return database.ArangoBackupStatus{
			Available:         true,
			ArangoBackupState: newState(database.ArangoBackupStateUpload, "", nil),
			Backup:            backup.Status.Backup.DeepCopy(),
		}, nil
	}

	return database.ArangoBackupStatus{
		Available:         true,
		ArangoBackupState: backup.Status.ArangoBackupState,
		Backup:            backup.Status.Backup.DeepCopy(),
	}, nil
}
