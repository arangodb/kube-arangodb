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
	"fmt"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

const (
	actionLocalJobID = "job-id"
)

func init() {
	registerAction(api.ActionTypeBackupRestore, newBackupRestoreAction, backupRestoreTimeout)
}

func newBackupRestoreAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionBackupRestore{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx)

	return a
}

// actionBackupRestore implements an BackupRestore.
type actionBackupRestore struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start runs restore operation.
func (a *actionBackupRestore) Start(ctx context.Context) (bool, error) {
	spec := a.actionCtx.GetSpec()
	status := a.actionCtx.GetStatusSnapshot()

	if spec.RestoreFrom == nil {
		return true, nil
	}

	if status.Restore != nil {
		a.log.Warn().Msg("Backup restore status should not be nil")
		return true, nil
	}

	backupResource, err := a.actionCtx.GetBackup(ctx, *spec.RestoreFrom)
	if err != nil {
		a.log.Error().Err(err).Msg("Unable to find backup")
		return true, nil
	}

	if backupResource.Status.Backup == nil {
		a.log.Error().Msg("Backup ID is not set")
		return true, nil
	}

	if err = a.updateRestoreStatus(ctx, api.DeploymentRestoreStateRestoring, ""); err != nil {
		// Try again in a while.
		return false, err
	}

	if spec.GetMode() == api.DeploymentModeCluster {
		// In a cluster mode a restore operation should be launched asynchronously.
		var op conn.ASyncFunc = func(ctx context.Context, cli driver.Client) error {
			return cli.Backup().Restore(ctx, driver.BackupID(backupResource.Status.Backup.ID), nil)
		}

		jobID, err := a.actionCtx.RunAsyncRequest(ctx, op)
		if err == nil {
			a.action = a.action.AddLocal(actionLocalJobID, jobID)
		} else {
			a.log.Error().Err(err).Msg("Restore failed")
		}

		return false, err
	}

	// In a Non-cluster mode a request should be sent synchronously.
	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	dbc, err := a.actionCtx.GetDatabaseClient(ctxChild)
	if err != nil {
		return false, err
	}

	restoreError := dbc.Backup().Restore(ctx, driver.BackupID(backupResource.Status.Backup.ID), nil)
	if restoreError == nil {
		if err = a.updateRestoreStatus(ctx, api.DeploymentRestoreStateRestored, ""); err != nil {
			// Next iteration will launch Restore again, and then it will try to save restore status.
			return false, err
		}

		return true, nil
	}

	a.log.Error().Err(restoreError).Msg("Restore failed")
	if err = a.updateRestoreStatus(ctx, api.DeploymentRestoreStateRestoreFailed, restoreError.Error()); err != nil {
		// Next iteration will launch Restore again, and then it will try to save restore status.
		return false, err
	}

	// When a Restore operation fails it must be considered as finished.
	return true, nil
}

// CheckProgress checks whether restore job is finished.
func (a *actionBackupRestore) CheckProgress(ctx context.Context) (bool, bool, error) {
	spec := a.actionCtx.GetSpec()
	if spec.GetMode() != api.DeploymentModeCluster {
		// Job ID can not be fetched in non-cluster mode.
		return true, false, nil
	}

	jobID, _ := a.action.GetLocal(actionLocalJobID)
	if len(jobID) == 0 {
		a.log.Error().Msg("Restoring backup is not possible. Job Id is empty in local action memory")
		return true, false, nil
	}
	log := a.log.With().Str(actionLocalJobID, jobID).Logger()

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	dbc, err := a.actionCtx.GetDatabaseClient(ctxChild)
	if err != nil {
		return false, false, errors.WithStack(err)
	}

	restoreError := globals.GetGlobalTimeouts().ArangoD().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return arangod.IsJobFinished(ctxChild, dbc, jobID)
	})

	if restoreError == nil {
		if err := a.updateRestoreStatus(ctx, api.DeploymentRestoreStateRestored, ""); err != nil {
			log.Error().Err(err).Msg("restore operation finished successfully, but status could not be updated")
			// fallthrough and return isReady=true because it is not possible to fetch job status anymore.
			// In this case restore status will be invalid.
		}

		return true, false, nil
	}

	if driver.IsNotFound(restoreError) {
		// Actually it is not known if a Restore job is finished.
		if err = a.updateRestoreStatus(ctx, api.DeploymentRestoreStateRestoreFailed, "restore job ID not found"); err != nil {
			log.Error().Err(err).Msg("restore operation is gone, and status could not be updated")
			// fallthrough and return isReady=true because this action is no longer valid.
			// In this case restore status will be invalid.
		}

		return true, false, nil
	}

	message := restoreError.Error()
	if ar, ok := driver.AsArangoError(err); ok && ar.Code == http.StatusNoContent {
		message = "restore job is pending or not finished yet"
	}

	if err = a.updateRestoreStatus(ctx, api.DeploymentRestoreStateRestoring, message); err != nil {
		log.Info().Err(err).Msg("restore operation is being restored, but status could not be updated")
		// fallthrough and return isReady=false.
		// Next iteration should reconcile it because it is possible to fetch job status again.
	}

	return false, false, err
}

// updateRestoreStatus updates current status of a restore action.
func (a *actionBackupRestore) updateRestoreStatus(ctx context.Context, state api.DeploymentRestoreState, message string) error {
	spec := a.actionCtx.GetSpec()

	err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		s.Restore = &api.DeploymentRestoreResult{
			RequestedFrom: spec.GetRestoreFrom(),
			State:         state,
			Message:       message,
		}

		return true
	})

	return errors.WithMessage(err, fmt.Sprintf("Unable setting restore state to \"%s\"", state))
}
