package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statePendingHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	_, err := h.getArangoDeploymentObject(backup)
	if err != nil {
		return createFailedState(err, backup.Status), nil
	}

	// Ensure that only specified number of processes are running
	backups, err := h.client.DatabaseV1alpha().ArangoBackups(backup.Namespace).List(meta.ListOptions{})
	if err != nil {
		return database.ArangoBackupStatus{}, err
	}

	count := 0
	for _, presentBackup := range backups.Items {
		if presentBackup.Name == backup.Name {
			break
		}

		if presentBackup.Spec.Deployment.Name != backup.Spec.Deployment.Name {
			break
		}

		count++
	}

	if count >= 1 {
		return database.ArangoBackupStatus{
			State:database.ArangoBackupState{
				State: database.ArangoBackupStatePending,
				Message: "backup already in process",
			},
		}, nil
	}

	return database.ArangoBackupStatus{
		State:database.ArangoBackupState{
			State:database.ArangoBackupStateScheduled,
		},
	}, nil
}

