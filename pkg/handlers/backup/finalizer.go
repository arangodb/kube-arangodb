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
	"context"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
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
		logger.Warn("After finalizing on object %s %s/%s finalizers left: %d",
			backup.GroupVersionKind().String(),
			backup.Namespace,
			backup.Name,
			i)
	}

	if _, err := h.client.BackupV1().ArangoBackups(backup.Namespace).Update(context.Background(), backup, meta.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (h *handler) finalizeBackup(backup *backupApi.ArangoBackup) error {
	lock := h.getDeploymentMutex(backup.Namespace, backup.Spec.Deployment.Name)
	lock.Lock()
	defer lock.Unlock()

	if backup.Status.Backup == nil {
		// No details passed, object can be removed
		return nil
	}

	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		// If deployment is not found we do not have to delete backup in database
		if apiErrors.IsNotFound(err) {
			return nil
		}

		if c, ok := err.(errors.Causer); ok {
			if apiErrors.IsNotFound(c.Cause()) {
				return nil
			}
		}

		return err
	}

	backups, err := h.client.BackupV1().ArangoBackups(backup.Namespace).List(context.Background(), meta.ListOptions{})
	if err != nil {
		return err
	}

	for _, existingBackup := range backups.Items {
		if existingBackup.Name == backup.Name {
			continue
		}

		if existingBackup.Status.Backup == nil {
			continue
		}

		// This backup is still in use
		if existingBackup.Status.Backup.ID == backup.Status.Backup.ID {
			return nil
		}
	}

	client, err := h.arangoClientFactory(deployment, backup)
	if err != nil {
		return err
	}

	if err = h.finalizeBackupAction(backup, client); err != nil {
		logger.Err(err).Warn("Operation abort failed for %s %s/%s",
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
		if !hasFinalizer(finalizer) {
			return false
		}
	}

	return true
}

func hasFinalizer(finalizer string) bool {
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
		if !hasFinalizer(finalizer) {
			old = append(old, finalizer)
		}
	}

	return old
}
