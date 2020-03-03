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
// Author Lars Maier
//

package backup

import (
	"context"

	"github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/rs/zerolog"
)

type Context interface {
	// GetSpec returns the current specification of the deployment
	GetSpec() api.DeploymentSpec
	// GetStatus returns the current status of the deployment
	GetStatus() (api.DeploymentStatus, int32)
	// UpdateStatus replaces the status of the deployment with the given status and
	// updates the resources in k8s.
	UpdateStatus(status api.DeploymentStatus, lastVersion int32, force ...bool) error
	// GetDatabaseClient returns a cached client for the entire database (cluster coordinators or single server),
	// creating one if needed.
	GetDatabaseClient(ctx context.Context) (driver.Client, error)
	// GetBackup receives information about a backup resource
	GetBackup(backup string) (*backupApi.ArangoBackup, error)
}

type BackupHandler struct {
	log     zerolog.Logger
	context Context
}

func NewHandler(log zerolog.Logger, context Context) *BackupHandler {
	return &BackupHandler{
		log:     log,
		context: context,
	}
}

func (b *BackupHandler) restoreFrom(backupName string) error {
	ctx := context.Background()
	dbc, err := b.context.GetDatabaseClient(ctx)
	if err != nil {
		return err
	}

	backupResource, err := b.context.GetBackup(backupName)
	if err != nil {
		return err
	}

	backupID := backupResource.Status.Backup.ID

	// trigger the actual restore
	if err := dbc.Backup().Restore(ctx, driver.BackupID(backupID), nil); err != nil {
		return err
	}

	return nil
}

func (b *BackupHandler) CheckRestore() error {

	spec := b.context.GetSpec()
	status, version := b.context.GetStatus()

	if spec.HasRestoreFrom() {
		// We have to trigger a restore operation
		if status.Restore == nil || status.Restore.RequestedFrom != spec.GetRestoreFrom() {
			// Prepare message that we are starting restore
			result := &api.DeploymentRestoreResult{
				RequestedFrom: spec.GetRestoreFrom(),
			}

			result.State = api.DeploymentRestoreStateRestoring

			for i := 0; i < 100; i++ {
				status, version := b.context.GetStatus()
				status.Restore = result
				b.context.UpdateStatus(status, version)
			}

			// Request restoring
			err := b.restoreFrom(spec.GetRestoreFrom())

			if err != nil {
				result.State = api.DeploymentRestoreStateRestoreFailed
				result.Message = err.Error()
			} else {
				result.State = api.DeploymentRestoreStateRestored
			}

			// try to update the status
			for i := 0; i < 100; i++ {
				status, version := b.context.GetStatus()
				status.Restore = result
				b.context.UpdateStatus(status, version)
			}
		}

		return nil
	}

	if status.Restore == nil {
		return nil
	}

	// Remove the restore entry from status
	status.Restore = nil
	b.context.UpdateStatus(status, version)
	return nil

}
