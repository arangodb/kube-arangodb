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
	"github.com/arangodb/go-driver"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/utils"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (h *handler) finalize(backup *backupApi.ArangoBackup) error {
	if backup.Finalizers == nil || len(backup.Finalizers) == 0 {
		return nil
	}

	finalizersToRemove := make(utils.StringList, len(backup.Finalizers))
	var finalizers utils.StringList = backup.Finalizers

	for _, finalizer := range finalizers {
		switch finalizer {
		case backupApi.FinalizerArangoBackup:
			if err := h.finalizeBackup(backup); err != nil {
				return err
			}
			finalizersToRemove = append(finalizersToRemove, backupApi.FinalizerArangoBackup)

			h.eventRecorder.Normal(backup, FinalizerChange, "Removed Finalizer: %s", backupApi.FinalizerArangoBackup)
		}
	}

	backup.Finalizers = finalizers.Remove(finalizersToRemove...)

	if i := len(backup.Finalizers); i > 0 {
		log.Warn().Msgf("After finalizing on object %s %s/%s finalizers left: %d",
			backup.GroupVersionKind().String(),
			backup.Namespace,
			backup.Name,
			i)
	}

	if _, err := h.client.BackupV1alpha().ArangoBackups(backup.Namespace).Update(backup); err != nil {
		return err
	}

	return nil
}

func (h *handler) finalizeBackup(backup *backupApi.ArangoBackup) error {
	if backup.Status.Backup == nil {
		// No details passed, object can be removed
		return nil
	}

	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		// If deployment is not found we do not have to delete backup in database
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	client, err := h.arangoClientFactory(deployment, backup)
	if err != nil {
		return err
	}

	if err = h.finalizeBackupAction(backup, client); err != nil {
		log.Warn().Err(err).Msgf("Operation abort failed for %s %s/%s",
			backup.GroupVersionKind().String(),
			backup.Namespace,
			backup.Name)
	}

	exists, err := client.Exists(driver.BackupID(backup.Status.Backup.ID))
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	err = client.Delete(driver.BackupID(backup.Status.Backup.ID))
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) finalizeBackupAction(backup *backupApi.ArangoBackup, client ArangoBackupClient) error {
	if backup.Status.Progress == nil {
		return nil
	}
	status, err := client.Progress(driver.BackupTransferJobID(backup.Status.Progress.JobID))
	if err != nil {
		return err
	}

	if status.Failed || status.Completed {
		return nil
	}

	if err = client.Abort(driver.BackupTransferJobID(backup.Status.Progress.JobID)); err != nil {
		return err
	}

	return nil
}

func hasFinalizers(backup *backupApi.ArangoBackup) bool {
	if backup.Finalizers == nil {
		return false
	}

	if len(backupApi.FinalizersArangoBackup) > len(backup.Finalizers) {
		return false
	}

	for _, finalizer := range backupApi.FinalizersArangoBackup {
		if !hasFinalizer(backup, finalizer) {
			return false
		}
	}

	return true
}

func hasFinalizer(backup *backupApi.ArangoBackup, finalizer string) bool {
	if backupApi.FinalizersArangoBackup == nil {
		return false
	}

	for _, existingFinalizer := range backupApi.FinalizersArangoBackup {
		if finalizer == existingFinalizer {
			return true
		}
	}

	return false
}

func appendFinalizers(backup *backupApi.ArangoBackup) []string {
	if backup.Finalizers == nil {
		return backupApi.FinalizersArangoBackup
	}

	if len(backup.Finalizers) == 0 {
		return backupApi.FinalizersArangoBackup
	}

	old := backup.Finalizers

	for _, finalizer := range backupApi.FinalizersArangoBackup {
		if !hasFinalizer(backup, finalizer) {
			old = append(old, finalizer)
		}
	}

	return old
}
