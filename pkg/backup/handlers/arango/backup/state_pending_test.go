package backup

import (
	"fmt"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_State_Pending_Common(t *testing.T) {
	wrapperUndefinedDeployment(t, database.ArangoBackupStatePending)
}

func Test_State_Pending_CheckNamespaceIsolation(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStatePending)
	deployment.Namespace = "non-existent"

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateFailed)

	require.Equal(t, newObj.Status.State.Message, fmt.Sprintf("%s \"%s\" not found", database.ArangoDeploymentCRDName, obj.Name))
}

func Test_State_Pending_OneBackupObject(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStatePending)

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateScheduled)

	require.False(t, newObj.Status.Available)
}

func Test_State_Pending_MultipleBackupObjectWithLimitation(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStatePending)
	obj2, _ := newObjectSet(database.ArangoBackupStatePending)
	obj2.Namespace = obj.Namespace
	obj2.Spec.Deployment.Name = deployment.Name

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj, obj2)

	t.Run("First backup object", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj)
		require.Equal(t, newObj.Status.State.State, database.ArangoBackupStateScheduled)

		require.False(t, newObj.Status.Available)
	})

	t.Run("Second backup object", func(t *testing.T) {
		require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj2)))

		// Assert
		newObj := refreshArangoBackup(t, handler, obj2)
		require.Equal(t, newObj.Status.State.State, database.ArangoBackupStatePending)
		require.Equal(t, newObj.Status.State.Message, "backup already in process")
	})
}