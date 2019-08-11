package backup

import (
	"fmt"
	"github.com/arangodb/go-driver"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

func stateDownloadHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
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

	meta, err := client.Get(driver.BackupID(backup.Status.Details.ID))
	if err != nil {
		return switchTemporaryError(err, backup.Status)
	}

	jobID, err := client.Download(meta.ID)
	if err != nil {
		return switchTemporaryError(err, backup.Status)
	}

	return database.ArangoBackupStatus{
		Available: false,
		State: database.ArangoBackupState{
			State: database.ArangoBackupStateDownloading,
			Progress: &database.ArangoBackupProgress{
				JobID:    string(jobID),
				Progress: "0%",
			},
		},
		Details: backup.Status.Details,
	}, nil
}