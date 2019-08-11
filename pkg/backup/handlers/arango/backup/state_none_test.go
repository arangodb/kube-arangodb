package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_State_None_Success(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, _ := newObjectSet(database.ArangoBackupStateNone)

	// Act
	createArangoBackup(t, handler, obj)
	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationUpdate, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status.State.State, database.ArangoBackupStatePending)

	require.False(t, newObj.Status.Available)
}
