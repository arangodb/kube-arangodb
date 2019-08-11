package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

func stateFailedHandler(h *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	return backup.Status, nil
}