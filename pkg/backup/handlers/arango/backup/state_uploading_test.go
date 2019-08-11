package backup

import (
	"fmt"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_State_Uploading_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, database.ArangoBackupStateUploading)
	wrapperConnectionIssues(t, database.ArangoBackupStateUploading)
	wrapperProgressMissing(t, database.ArangoBackupStateUploading)
}

func Test_State_Uploading_Success(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateUploading)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Version:backupMeta.Version,
		CreationTimestamp:now(),
	}

	obj.Status.State.Progress = &database.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("Restore percent", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateUploading, newObj.Status.State.State)
		require.Equal(t, fmt.Sprintf("%d%%",0), newObj.Status.State.Progress.Progress)
		require.Equal(t, obj.Status.State.Progress.JobID, newObj.Status.State.Progress.JobID)

		require.True(t, newObj.Status.Available)
	})

	t.Run("Restore percent after update", func(t *testing.T) {
		p := 55
		mock.progresses[progress] = ArangoBackupProgress{
			Progress: p,
		}

		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateUploading, newObj.Status.State.State)
		require.Equal(t, fmt.Sprintf("%d%%",p), newObj.Status.State.Progress.Progress)
		require.Equal(t, fmt.Sprintf("%s",progress), newObj.Status.State.Progress.JobID)

		require.True(t, newObj.Status.Available)
	})

	t.Run("Finished", func(t *testing.T) {
		mock.progresses[progress] = ArangoBackupProgress{
			Completed: true,
		}

		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateReady, newObj.Status.State.State)
		require.Nil(t, newObj.Status.State.Progress)

		require.True(t, newObj.Status.Available)
	})
}

func Test_State_Uploading_FailedUpload(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateUploading)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	errorMsg := "error"
	mock.progresses[progress] = ArangoBackupProgress{
		Failed: true,
		FailMessage: errorMsg,
	}

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}

	obj.Status.State.Progress = &database.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateFailed, newObj.Status.State.State)
	require.Equal(t, fmt.Sprintf("upload failed with error: %s", errorMsg), newObj.Status.State.Message)
	require.Nil(t, newObj.Status.State.Progress)

	require.False(t, newObj.Status.Available)
}

func Test_State_Uploading_FailedProgress(t *testing.T) {
	// Arrange
	errorMsg := "progress error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		progressError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateUploading)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}

	obj.Status.State.Progress = &database.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, database.ArangoBackupStateFailed, newObj.Status.State.State)
	require.Equal(t, errorMsg, newObj.Status.State.Message)
	require.Nil(t, newObj.Status.State.Progress)

	require.False(t, newObj.Status.Available)
}

func Test_State_Uploading_TemporaryFailedProgress(t *testing.T) {
	// Arrange
	errorMsg := "progress error"
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{
		isTemporaryError: true,
		progressError: errorMsg,
	})

	obj, deployment := newObjectSet(database.ArangoBackupStateUploading)

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	progress, err := mock.Upload(backupMeta.ID)
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:                string(backupMeta.ID),
		Version:           backupMeta.Version,
		CreationTimestamp: now(),
	}

	obj.Status.State.Progress = &database.ArangoBackupProgress{
		JobID: string(progress),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	err = handler.Handle(newItemFromBackup(operator.OperationUpdate, obj))

	// Assert
	compareTemporaryState(t, err, errorMsg, handler, obj)
}