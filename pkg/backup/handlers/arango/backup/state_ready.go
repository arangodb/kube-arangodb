package backup

import (
	"fmt"
	"github.com/arangodb/go-driver"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

func stateReadyHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	deployment, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return createFailedState(err, backup.Status), nil
	}

	client, err := h.arangoClientFactory(deployment)
	if err != nil {
		return database.ArangoBackupStatus{}, NewTemporaryError("unable to create client: %s", err.Error())
	}

	if backup.Status.Details == nil {
		return createFailedState(fmt.Errorf("backup details are missing"), backup.Status), nil
	}

	_, err = client.Get(driver.BackupID(backup.Status.Details.ID))
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

	return database.ArangoBackupStatus{
		Available: true,
		State: database.ArangoBackupState{
			State: database.ArangoBackupStateReady,
		},
		Details: backup.Status.Details,
	}, nil
}