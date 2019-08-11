package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_State_Download_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, database.ArangoBackupStateDownload)
	wrapperConnectionIssues(t, database.ArangoBackupStateDownload)
}

func Test_State_Download_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateDownload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Version:backupMeta.Version,
		CreationTimestamp:now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateDownloading)

	require.NotNil(t, newObj.Status.State.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 1)
	require.Equal(t, progresses[0], newObj.Status.State.Progress.JobID)

	require.False(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Details)
	require.Equal(t, string(backupMeta.ID), newObj.Status.Details.ID)
	require.Equal(t, backupMeta.Version, newObj.Status.Details.Version)
}

func Test_State_Download_GetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		getError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateDownload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Version:backupMeta.Version,
		CreationTimestamp:now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateFailed)

	require.Nil(t, newObj.Status.State.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 0)

	require.False(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Details)
	require.Equal(t, string(backupMeta.ID), newObj.Status.Details.ID)
	require.Equal(t, backupMeta.Version, newObj.Status.Details.Version)
}

func Test_State_Download_TemporaryGetFailed(t *testing.T) {
	// Arrange
	errorMsg := "get error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError:true,

		getError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateDownload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Version:backupMeta.Version,
		CreationTimestamp:now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err = handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}

func Test_State_Download_DownloadFailed(t *testing.T) {
	// Arrange
	errorMsg := "download error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		downloadError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateDownload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Version:backupMeta.Version,
		CreationTimestamp:now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateFailed)

	require.Nil(t, newObj.Status.State.Progress)
	progresses := mock.getProgressIDs()
	require.Len(t, progresses, 0)

	require.False(t, newObj.Status.Available)

	require.NotNil(t, newObj.Status.Details)
	require.Equal(t, string(backupMeta.ID), newObj.Status.Details.ID)
	require.Equal(t, backupMeta.Version, newObj.Status.Details.Version)
}

func Test_State_Download_TemporaryDownloadFailed(t *testing.T) {
	// Arrange
	errorMsg := "download error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError:true,
		downloadError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateDownload)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Version:backupMeta.Version,
		CreationTimestamp:now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err = handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}