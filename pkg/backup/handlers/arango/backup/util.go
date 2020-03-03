//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
	"strings"

	clientBackup "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/backup/v1"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/backup/state"
)

var (
	progressStates = []state.State{
		backupApi.ArangoBackupStateScheduled,
		backupApi.ArangoBackupStateCreate,
		backupApi.ArangoBackupStateDownload,
		backupApi.ArangoBackupStateDownloading,
		backupApi.ArangoBackupStateUpload,
		backupApi.ArangoBackupStateUploading,
	}
)

func inProgress(backup *backupApi.ArangoBackup) bool {
	for _, state := range progressStates {
		if state == backup.Status.State {
			return true
		}
	}

	return false
}

func isBackupRunning(backup *backupApi.ArangoBackup, client clientBackup.ArangoBackupInterface) (bool, error) {
	backups, err := client.List(meta.ListOptions{})

	if err != nil {
		return false, newTemporaryError(err)
	}

	for _, existingBackup := range backups.Items {
		if existingBackup.Name == backup.Name {
			continue
		}

		// We can upload multiple uploads from same deployment in same time
		if backup.Status.State == backupApi.ArangoBackupStateReady &&
			(existingBackup.Status.State == backupApi.ArangoBackupStateUpload || existingBackup.Status.State == backupApi.ArangoBackupStateUploading) {
			if backupUpload := backup.Status.Backup; backupUpload != nil {
				if existingBackupUpload := existingBackup.Status.Backup; existingBackupUpload != nil {
					if strings.ToLower(backupUpload.ID) == strings.ToLower(existingBackupUpload.ID) {
						return true, nil
					}
				}
			}
		} else {
			if existingBackup.Spec.Deployment.Name != backup.Spec.Deployment.Name {
				continue
			}

			if inProgress(&existingBackup) {
				return true, nil
			}
		}
	}

	return false, nil
}
