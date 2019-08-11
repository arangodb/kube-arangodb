package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func switchTemporaryError(err error, status database.ArangoBackupStatus) (database.ArangoBackupStatus, error) {
	if IsTemporaryError(err) {
		return database.ArangoBackupStatus{}, err
	}

	return createFailedState(err, status), nil
}

func createFailedState(err error, status database.ArangoBackupStatus) database.ArangoBackupStatus {
	newStatus := status.DeepCopy()

	newStatus.State = database.ArangoBackupState{
		State:database.ArangoBackupStateFailed,
		Message:err.Error(),
	}

	newStatus.Available = false

	return *newStatus
}

func now() meta.Timestamp {
	t := meta.Now()

	timestamp := (&t).ProtoTime()

	return *timestamp
}