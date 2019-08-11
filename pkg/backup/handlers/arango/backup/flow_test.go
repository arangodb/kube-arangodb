package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Flow_SuccessHappyPath(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateNone)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	t.Run("Change from None to Pending", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStatePending, newObj.Status.State.State)
	})

	t.Run("Change from Pending to Scheduled", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateScheduled, newObj.Status.State.State)
	})

	t.Run("Change from Scheduled to Create", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateCreate, newObj.Status.State.State)
	})

	t.Run("Change from Create to Ready", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateReady, newObj.Status.State.State)
	})

	t.Run("Ensure Ready State Keeps", func(t *testing.T) {
		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateReady, newObj.Status.State.State)
	})

	t.Run("Change from Ready to Deleted", func(t *testing.T) {
		// Arrange
		mock.errors.getError = "error"

		// Act
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, database.ArangoBackupStateDeleted, newObj.Status.State.State)
	})
}
