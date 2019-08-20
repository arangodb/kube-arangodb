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
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/utils"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (h *handler) finalize(backup *database.ArangoBackup) error {
	if backup.Finalizers == nil || len(backup.Finalizers) == 0 {
		return nil
	}

	finalizersToRemove := make(utils.StringList, len(backup.Finalizers))
	var finalizers utils.StringList = backup.Finalizers

	for _, finalizer := range finalizers {
		switch finalizer {
		case database.FinalizerArangoBackup:
			if err := h.finalizeBackup(backup); err != nil {
				return err
			}
			finalizersToRemove = append(finalizersToRemove, database.FinalizerArangoBackup)

			h.eventRecorder.Normal(backup, FinalizerChange, "Removed Finalizer: %s", database.FinalizerArangoBackup)
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

	if _, err := h.client.DatabaseV1alpha().ArangoBackups(backup.Namespace).Update(backup); err != nil {
		return err
	}

	return nil
}

func (h *handler) finalizeBackup(backup *database.ArangoBackup) error {
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

func (h *handler) finalizeBackupAction(backup *database.ArangoBackup, client ArangoBackupClient) error {
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

func hasFinalizers(backup *database.ArangoBackup) bool {
	if backup.Finalizers == nil {
		return false
	}

	if len(database.FinalizersArangoBackup) > len(backup.Finalizers) {
		return false
	}

	for _, finalizer := range database.FinalizersArangoBackup {
		if !hasFinalizer(backup, finalizer) {
			return false
		}
	}

	return true
}

func hasFinalizer(backup *database.ArangoBackup, finalizer string) bool {
	if backup.Finalizers == nil {
		return false
	}

	for _, existingFinalizer := range backup.Finalizers {
		if finalizer == existingFinalizer {
			return true
		}
	}

	return false
}

func appendFinalizers(backup *database.ArangoBackup) []string {
	if backup.Finalizers == nil {
		return database.FinalizersArangoBackup
	}

	if len(backup.Finalizers) == 0 {
		return database.FinalizersArangoBackup
	}

	old := backup.Finalizers

	for _, finalizer := range database.FinalizersArangoBackup {
		if !hasFinalizer(backup, finalizer) {
			old = append(old, finalizer)
		}
	}

	return old
}
