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
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/backup/state"
	"github.com/arangodb/kube-arangodb/pkg/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type updateStatusFunc func(status *backupApi.ArangoBackupStatus)

func wrapUpdateStatus(backup *backupApi.ArangoBackup, update ...updateStatusFunc) (*backupApi.ArangoBackupStatus, error) {
	return updateStatus(backup, update...), nil
}

func updateStatus(backup *backupApi.ArangoBackup, update ...updateStatusFunc) *backupApi.ArangoBackupStatus {
	s := backup.Status.DeepCopy()

	for _, u := range update {
		u(s)
	}

	return s
}

func updateStatusState(state state.State, template string, a ...interface{}) updateStatusFunc {
	return func(status *backupApi.ArangoBackupStatus) {
		if status.State != state {
			status.Time = v1.Now()
		}
		status.State = state
		status.Message = fmt.Sprintf(template, a...)
	}
}

func updateStatusAvailable(available bool) updateStatusFunc {
	return func(status *backupApi.ArangoBackupStatus) {
		status.Available = available
	}
}

func updateStatusJob(id, progress string) updateStatusFunc {
	return func(status *backupApi.ArangoBackupStatus) {
		status.Progress = &backupApi.ArangoBackupProgress{
			JobID:    id,
			Progress: progress,
		}
	}
}

func updateStatusBackupUpload(uploaded *bool) updateStatusFunc {
	return func(status *backupApi.ArangoBackupStatus) {
		if status.Backup != nil {
			status.Backup.Uploaded = uploaded
		}
	}
}

func updateStatusBackupImported(imported *bool) updateStatusFunc {
	return func(status *backupApi.ArangoBackupStatus) {
		if status.Backup != nil {
			status.Backup.Imported = imported
		}
	}
}

func updateStatusBackupDownload(downloaded *bool) updateStatusFunc {
	return func(status *backupApi.ArangoBackupStatus) {
		if status.Backup != nil {
			status.Backup.Downloaded = downloaded
		}
	}
}

func updateStatusBackup(backupMeta driver.BackupMeta) updateStatusFunc {
	return func(status *backupApi.ArangoBackupStatus) {
		status.Backup = createBackupFromMeta(backupMeta, status.Backup)
	}
}

func cleanStatusJob() updateStatusFunc {
	return func(status *backupApi.ArangoBackupStatus) {
		status.Progress = nil
	}
}

func setFailedState(backup *backupApi.ArangoBackup, err error) (*backupApi.ArangoBackupStatus, error) {
	return wrapUpdateStatus(backup,
		updateStatusState(backupApi.ArangoBackupStateFailed, createStateMessage(backup.Status.State, backupApi.ArangoBackupStateFailed, err.Error())),
		updateStatusAvailable(false))
}

func createStateMessage(from, to state.State, message string) string {
	return fmt.Sprintf("Transiting from %s to %s: %s", from, to, message)
}

func switchTemporaryError(backup *backupApi.ArangoBackup, err error) (*backupApi.ArangoBackupStatus, error) {
	if _, ok := err.(temporaryError); ok {
		return nil, err
	}

	return setFailedState(backup, err)
}

func createBackupFromMeta(backupMeta driver.BackupMeta, old *backupApi.ArangoBackupDetails) *backupApi.ArangoBackupDetails {
	var obj *backupApi.ArangoBackupDetails

	if old == nil {
		obj = &backupApi.ArangoBackupDetails{}
	} else {
		obj = old.DeepCopy()
	}

	obj.PotentiallyInconsistent = util.NewBool(backupMeta.PotentiallyInconsistent)
	obj.SizeInBytes = backupMeta.SizeInBytes
	obj.CreationTimestamp = v1.Time{
		Time: backupMeta.DateTime,
	}
	obj.NumberOfDBServers = backupMeta.NumberOfDBServers
	obj.Version = backupMeta.Version
	obj.ID = string(backupMeta.ID)

	return obj
}
