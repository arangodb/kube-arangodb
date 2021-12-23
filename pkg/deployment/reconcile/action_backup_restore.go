//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	"github.com/arangodb/go-driver"
	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func init() {
	registerAction(api.ActionTypeBackupRestore, newBackupRestoreAction)
}

func newBackupRestoreAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionBackupRestore{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, backupRestoreTimeout)

	return a
}

// actionBackupRestore implements an BackupRestore.
type actionBackupRestore struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

func (a actionBackupRestore) Start(ctx context.Context) (bool, error) {
	spec := a.actionCtx.GetSpec()
	status := a.actionCtx.GetStatusSnapshot()

	if spec.RestoreFrom == nil {
		return true, nil
	}

	if status.Restore != nil {
		a.log.Warn().Msg("Backup restore status should not be nil")
		return true, nil
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	dbc, err := a.actionCtx.GetDatabaseClient(ctxChild)
	if err != nil {
		return false, err
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

	if err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		result := &api.DeploymentRestoreResult{
			RequestedFrom: spec.GetRestoreFrom(),
		}

		result.State = api.DeploymentRestoreStateRestoring

		s.Restore = result

		return true
	}, true); err != nil {
		return false, err
	}

	// The below action can take a while so the full parent timeout context is used.
	restoreError := dbc.Backup().Restore(ctx, driver.BackupID(backupResource.Status.Backup.ID), nil)
	if restoreError != nil {
		a.log.Error().Err(restoreError).Msg("Restore failed")
	}

	if err := a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		result := &api.DeploymentRestoreResult{
			RequestedFrom: spec.GetRestoreFrom(),
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
		a.log.Error().Err(err).Msg("Unable to ser restored state")
		return false, err
	}

	return true, nil
}
