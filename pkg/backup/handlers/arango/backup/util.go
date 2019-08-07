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
	"reflect"
	"strings"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/backup/utils"
	"github.com/rs/zerolog/log"

	clientBackup "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/backup/v1alpha"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
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

func switchTemporaryError(err error, status backupApi.ArangoBackupStatus) (backupApi.ArangoBackupStatus, error) {
	if checkTemporaryError(err) {
		return backupApi.ArangoBackupStatus{}, err
	}

	return createFailedState(err, status), nil
}

func createFailMessage(state state.State, message string) string {
	return fmt.Sprintf("Failed State %s: %s", state, message)
}

func createFailedState(err error, status backupApi.ArangoBackupStatus) backupApi.ArangoBackupStatus {
	e := log.Error().Err(err).Str("type", reflect.TypeOf(err).String())
	if c, ok := err.(utils.Causer); ok {
		e = e.AnErr("caused", c.Cause()).Str("causedType", reflect.TypeOf(c.Cause()).String()).Str("causedError", fmt.Sprintf("%v", c.Cause()))

		if a, ok := c.Cause().(driver.ArangoError); ok {
			e = e.Str("aMsg", a.ErrorMessage).Int("aCode", a.Code).Int("aNum", a.ErrorNum).Str("aMsg", a.ErrorMessage).Bool("aTemp", a.Temporary())
		}
	}
	e.Msgf("Error %v", err)

	newStatus := status.DeepCopy()

	newStatus.ArangoBackupState = newState(backupApi.ArangoBackupStateFailed, createFailMessage(status.State, err.Error()), nil)

	newStatus.Available = false

	return *newStatus
}

func newState(state state.State, message string, progress *backupApi.ArangoBackupProgress) backupApi.ArangoBackupState {
	return backupApi.ArangoBackupState{
		State: state,
		Time:  meta.Now(),

		Message: message,

		Progress: progress,
	}
}

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
		return false, err
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
