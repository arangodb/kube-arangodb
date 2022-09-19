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

package reconcile

import (
	"context"
	"time"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

const (
	actionBackupRestoreLocalJobID      api.PlanLocalKey = "jobID"
	actionBackupRestoreLocalBackupName api.PlanLocalKey = "backupName"
)

func newBackupRestoreAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionBackupRestore{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionBackupRestore implements an BackupRestore.
type actionBackupRestore struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

func (a actionBackupRestore) Start(ctx context.Context) (bool, error) {
	spec := a.actionCtx.GetSpec()
	status := a.actionCtx.GetStatus()

	if spec.RestoreFrom == nil {
		return true, nil
	}

	if status.Restore != nil {
		a.log.Warn("Backup restore status should not be nil")
		return true, nil
	}

	backupResource, err := a.actionCtx.GetBackup(ctx, *spec.RestoreFrom)
	if err != nil {
		a.log.Err(err).Error("Unable to find backup")
		return true, nil
	}

	if backupResource.Status.Backup == nil {
		a.log.Error("Backup ID is not set")
		return true, nil
	}

	if err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		result := &api.DeploymentRestoreResult{
			RequestedFrom: spec.GetRestoreFrom(),
		}

		result.State = api.DeploymentRestoreStateRestoring

		s.Restore = result

		return true
	}); err != nil {
		return false, err
	}

	switch mode := a.actionCtx.GetSpec().Mode.Get(); mode {
	case api.DeploymentModeActiveFailover, api.DeploymentModeSingle:
		return a.restoreSync(ctx, backupResource)
	case api.DeploymentModeCluster:
		return a.restoreAsync(ctx, backupResource)
	default:
		return false, errors.Newf("Unknown mode %s", mode)
	}
}

func (a actionBackupRestore) restoreAsync(ctx context.Context, backup *backupApi.ArangoBackup) (bool, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	dbc, err := a.actionCtx.GetDatabaseAsyncClient(ctxChild)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to create client")
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	if err := dbc.Backup().Restore(ctxChild, driver.BackupID(backup.Status.Backup.ID), nil); err != nil {
		if id, ok := conn.IsAsyncJobInProgress(err); ok {
			a.actionCtx.Add(actionBackupRestoreLocalJobID, id, true)
			a.actionCtx.Add(actionBackupRestoreLocalBackupName, backup.GetName(), true)

			// Async request has been send
			return false, nil
		} else {
			return false, errors.Wrapf(err, "Unknown restore error")
		}
	}

	return false, errors.Newf("Async response not received")
}

func (a actionBackupRestore) restoreSync(ctx context.Context, backup *backupApi.ArangoBackup) (bool, error) {
	dbc, err := a.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		a.log.Err(err).Debug("Failed to create database client")
		return false, nil
	}

	// The below action can take a while so the full parent timeout context is used.
	restoreError := dbc.Backup().Restore(ctx, driver.BackupID(backup.Status.Backup.ID), nil)
	if restoreError != nil {
		a.log.Err(restoreError).Error("Restore failed")
	}

	if err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		result := &api.DeploymentRestoreResult{
			RequestedFrom: backup.GetName(),
		}

		if restoreError != nil {
			result.State = api.DeploymentRestoreStateRestoreFailed
			result.Message = restoreError.Error()
		} else {
			result.State = api.DeploymentRestoreStateRestored
		}

		s.Restore = result

		return true
	}); err != nil {
		a.log.Err(err).Error("Unable to set restored state")
		return false, err
	}

	return true, nil
}

func (a actionBackupRestore) CheckProgress(ctx context.Context) (bool, bool, error) {
	backup, ok := a.actionCtx.Get(a.action, actionBackupRestoreLocalBackupName)
	if !ok {
		return false, false, errors.Newf("Local Key is missing in action: %s", actionBackupRestoreLocalBackupName)
	}

	job, ok := a.actionCtx.Get(a.action, actionBackupRestoreLocalJobID)
	if !ok {
		return false, false, errors.Newf("Local Key is missing in action: %s", actionBackupRestoreLocalJobID)
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	dbc, err := a.actionCtx.GetDatabaseAsyncClient(ctxChild)
	if err != nil {
		a.log.Err(err).Debug("Failed to create database client")
		return false, false, nil
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	// Params does not matter in async fetch
	restoreError := dbc.Backup().Restore(conn.WithAsyncID(ctxChild, job), "", nil)
	if restoreError != nil {
		if _, ok := conn.IsAsyncJobInProgress(restoreError); ok {
			// Job still in progress
			return false, false, nil
		}

		if errors.IsTemporary(restoreError) {
			// Retry
			return false, false, nil
		}

		// Add wait grace period for restore jobs - async job creation is asynchronous
		if ok := conn.IsAsyncErrorNotFound(restoreError); ok {
			if s := a.action.StartTime; s != nil && !s.Time.IsZero() {
				if time.Since(s.Time) < 10*time.Second {
					// Retry
					return false, false, nil
				}
			}
		}
	}

	// Restore is done

	if err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		result := &api.DeploymentRestoreResult{
			RequestedFrom: backup,
			State:         api.DeploymentRestoreStateRestored,
		}

		if restoreError != nil {
			a.log.Err(restoreError).Error("Restore failed")
			result.State = api.DeploymentRestoreStateRestoreFailed
			result.Message = restoreError.Error()
		}

		s.Restore = result

		return true
	}); err != nil {
		a.log.Err(err).Error("Unable to set restored state")
		return false, false, err
	}

	return true, false, nil
}
