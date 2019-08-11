package backup

import (
	"fmt"
	"github.com/arangodb/go-driver"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

func stateUploadingHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
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

	if backup.Status.State.Progress == nil {
		return createFailedState(fmt.Errorf("backup progress details are missing"), backup.Status), nil
	}

	details, err := client.Progress(driver.BackupTransferJobID(backup.Status.State.Progress.JobID))
	if err != nil {
		return switchTemporaryError(err, backup.Status)
	}

	if details.Failed {
		return createFailedState(fmt.Errorf("upload failed with error: %s", details.FailMessage), backup.Status), nil
	}

	if details.Completed {
		return database.ArangoBackupStatus{
			Available: true,
			State: database.ArangoBackupState{
				State: database.ArangoBackupStateReady,
			},
			Details: backup.Status.Details,
		}, nil
	}

	return database.ArangoBackupStatus{
		Available: true,
		State: database.ArangoBackupState{
			State: database.ArangoBackupStateUploading,
			Progress: &database.ArangoBackupProgress{
				JobID:    backup.Status.State.Progress.JobID,
				Progress: fmt.Sprintf("%d%%", details.Progress),
			},
		},
		Details: backup.Status.Details,
	}, nil
}