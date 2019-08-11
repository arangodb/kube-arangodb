package backup

import (
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/state"
)

type stateHolder func(handler *handler, backup *database.ArangoBackup) (database.ArangoBackupStatus, error)

var(
	stateHolders = map[state.State]stateHolder {
		database.ArangoBackupStateNone: stateNoneHandler,
		database.ArangoBackupStatePending: statePendingHandler,
		database.ArangoBackupStateScheduled: stateScheduledHandler,
		database.ArangoBackupStateCreate: stateCreateHandler,
		database.ArangoBackupStateUpload: stateUploadHandler,
		database.ArangoBackupStateUploading: stateUploadingHandler,
		database.ArangoBackupStateDownload: stateDownloadHandler,
		database.ArangoBackupStateDownloading: stateDownloadingHandler,
		database.ArangoBackupStateReady: stateReadyHandler,
		database.ArangoBackupStateDeleted: stateDeletedHandler,
		database.ArangoBackupStateFailed: stateFailedHandler,
}
)
