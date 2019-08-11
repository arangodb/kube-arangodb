package backup

import (
	"github.com/arangodb/go-driver"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

func stateCreateHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return createFailedState(err, backup.Status), nil
	}

	client, err := h.arangoClientFactory(deployment)
	if err != nil {
		return database.ArangoBackupStatus{}, NewTemporaryError("unable to create client: %s", err.Error())
	}

	var backupMeta driver.BackupMeta

	// Try to recover old backup. If old backup is missing go into deleted state
	if backup.Status.Details != nil {
		backupMeta, err = client.Get(driver.BackupID(backup.Status.Details.ID))
		if err != nil {
			if IsTemporaryError(err) {
				return switchTemporaryError(err, backup.Status)
			}

			// Go into deleted state
			return database.ArangoBackupStatus{
				State: database.ArangoBackupState{
					State: database.ArangoBackupStateDeleted,
				},
				Details: backup.Status.Details,
			}, nil
		}
	} else {
		backupMeta, err = client.Create()
		if err != nil {
			return switchTemporaryError(err, backup.Status)
		}
	}

	details := &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}

	if backup.Spec.Upload != nil {
		return database.ArangoBackupStatus{
			Available:true,
			State: database.ArangoBackupState{
				State: database.ArangoBackupStateUpload,
			},
			Details: details,
		}, nil
	}

	return database.ArangoBackupStatus{
		Available:true,
		State: database.ArangoBackupState{
			State: database.ArangoBackupStateReady,
		},
		Details: details,
	}, nil
}