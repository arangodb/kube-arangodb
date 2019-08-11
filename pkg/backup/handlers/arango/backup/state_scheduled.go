package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

func stateScheduledHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	_, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return createFailedState(err, backup.Status), nil
	}

	if backup.Spec.Download != nil {
		return database.ArangoBackupStatus{
			State:database.ArangoBackupState{
				State:database.ArangoBackupStateDownload,
			},
		}, nil
	}

	return database.ArangoBackupStatus{
		State:database.ArangoBackupState{
			State:database.ArangoBackupStateCreate,
		},
	}, nil
}