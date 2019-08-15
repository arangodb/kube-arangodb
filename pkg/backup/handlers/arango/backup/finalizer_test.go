package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func Test_Finalizer_PassThru(t *testing.T) {
	// Arrange
	handler, _ := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, _ := newObjectSet(database.ArangoBackupStateCreate)
	time := meta.Time{
		Time: time.Now(),
	}
	obj.DeletionTimestamp = &time

	// Act
	//createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)
}

func Test_Finalizer_RemoveObject(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Finalizers = []string{
		database.FinalizerArangoBackup,
	}

	time := meta.Now()
	obj.DeletionTimestamp = &time

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Forced:&backupMeta.Forced,
		Version:backupMeta.Version,
		CreationTimestamp:meta.Now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)

	require.Len(t, newObj.Finalizers, 0)

	exists, err := mock.Exists(backupMeta.ID)
	require.NoError(t, err)
	require.False(t, exists)
}

func Test_Finalizer_RemoveObject_WithoutFinalizer(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)

	time := meta.Now()
	obj.DeletionTimestamp = &time

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Forced:&backupMeta.Forced,
		Version:backupMeta.Version,
		CreationTimestamp:meta.Now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)

	require.Len(t, newObj.Finalizers, 0)

	exists, err := mock.Exists(backupMeta.ID)
	require.NoError(t, err)
	require.True(t, exists)
}

func Test_Finalizer_RemoveObject_UnknownFinalizer(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Finalizers = []string{
		"UNKNOWN",
	}

	time := meta.Now()
	obj.DeletionTimestamp = &time

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Forced:&backupMeta.Forced,
		Version:backupMeta.Version,
		CreationTimestamp:meta.Now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)

	require.Len(t, newObj.Finalizers, 1)

	exists, err := mock.Exists(backupMeta.ID)
	require.NoError(t, err)
	require.True(t, exists)
}

func Test_Finalizer_RemoveObject_MixedFinalizers(t *testing.T) {
	// Arrange
	handler, mock := newErrorsFakeHandler(mockErrorsArangoClientBackup{})

	obj, deployment := newObjectSet(database.ArangoBackupStateReady)
	obj.Finalizers = []string{
		"UNKNOWN",
		database.FinalizerArangoBackup,
	}

	time := meta.Now()
	obj.DeletionTimestamp = &time

	backupMeta, err := mock.Create()
	require.NoError(t, err)

	obj.Status.Details = &database.ArangoBackupDetails{
		ID:string(backupMeta.ID),
		Forced:&backupMeta.Forced,
		Version:backupMeta.Version,
		CreationTimestamp:meta.Now(),
	}

	// Act
	createArangoDeployment(t, handler, deployment)
	createArangoBackup(t, handler, obj)

	require.NoError(t, handler.Handle(newItemFromBackup(operator.OperationDelete, obj)))

	// Assert
	newObj := refreshArangoBackup(t, handler, obj)
	require.Equal(t, newObj.Status, obj.Status)
	require.Equal(t, newObj.Spec, obj.Spec)

	require.Len(t, newObj.Finalizers, 1)

	exists, err := mock.Exists(backupMeta.ID)
	require.NoError(t, err)
	require.False(t, exists)
}