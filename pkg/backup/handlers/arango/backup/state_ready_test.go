package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_State_Ready_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, database.ArangoBackupStateReady)
	wrapperConnectionIssues(t, database.ArangoBackupStateReady)
}

func Test_State_Ready_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, obj.Status, newObj.Status)
	require.True(t, newObj.Status.Available)
}

func Test_State_Ready_GetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError:errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateDeleted, newObj.Status.State.State)
	require.False(t, newObj.Status.Available)
}

func Test_State_Ready_TemporaryGetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError:true,
		getError:errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}
	obj.Status.Available = true

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err = handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}