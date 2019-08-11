package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_State_Create_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, database.ArangoBackupStateCreate)
	wrapperConnectionIssues(t, database.ArangoBackupStateCreate)
}

func Test_State_Create_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateReady)

	require.NotNil(t, newObj.Status.Details)

	backups := mock.getIDs()
	require.Len(t, backups, 1)

	require.Equal(t, newObj.Status.Details.ID, backups[0])
	require.Equal(t, newObj.Status.Details.Version, mockVersion)
}

func Test_State_Create_Upload(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateCreate)
	obj.Spec.Upload = &database.ArangoBackupSpecOperation{}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateUpload)

	require.NotNil(t, newObj.Status.Details)

	backups := mock.getIDs()
	require.Len(t, backups, 1)

	require.Equal(t, newObj.Status.Details.ID, backups[0])
	require.Equal(t, newObj.Status.Details.Version, mockVersion)

	require.True(t, newObj.Status.Available)
}

func Test_State_Create_CreateFailed(t *testing.T) {
	// Arrange
	errorMsg := "create error"
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		createError:errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateFailed)
	require.Equal(t, newObj.Status.State.Message, errorMsg)

	require.Nil(t, newObj.Status.Details)

	require.False(t, newObj.Status.Available)
}

func Test_State_Create_TemporaryCreateFailed(t *testing.T) {
	// Arrange
	errorMsg := "create error"
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError: true,
		createError:errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateCreate)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err := handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}

func Test_State_Create_GetFailedWithExistingDeploymentSpec(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError:errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateCreate)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID: "non-existent",
		Version: "non-existent",
		CreationTimestamp:now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateDeleted)

	require.NotNil(t, newObj.Status.Details)

	require.False(t, newObj.Status.Available)
}

func Test_State_Create_TemporaryGetFailedWithExistingDeploymentSpec(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError: true,
		getError:errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateCreate)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID: "non-existent",
		Version: "non-existent",
		CreationTimestamp:now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err := handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}