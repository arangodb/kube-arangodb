package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

func stateNoneHandler(*handler, *database.ArangoBackup) (database.ArangoBackupStatus, error) {
	return database.ArangoBackupStatus{
		State: database.ArangoBackupState{
			State: database.ArangoBackupStatePending,
		},
	}, nil
}
