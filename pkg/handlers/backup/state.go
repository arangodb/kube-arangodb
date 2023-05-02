//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package backup

import (
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/handlers/backup/state"
)

type stateHolder func(handler *handler, backup *backupApi.ArangoBackup) (*backupApi.ArangoBackupStatus, error)

var (
	stateHolders = map[state.State]stateHolder{
		backupApi.ArangoBackupStateNone:          stateNoneHandler,
		backupApi.ArangoBackupStatePending:       statePendingHandler,
		backupApi.ArangoBackupStateScheduled:     stateScheduledHandler,
		backupApi.ArangoBackupStateCreate:        stateCreateHandler,
		backupApi.ArangoBackupStateCreateError:   stateCreateErrorHandler,
		backupApi.ArangoBackupStateUpload:        stateUploadHandler,
		backupApi.ArangoBackupStateUploading:     stateUploadingHandler,
		backupApi.ArangoBackupStateUploadError:   stateUploadErrorHandler,
		backupApi.ArangoBackupStateDownload:      stateDownloadHandler,
		backupApi.ArangoBackupStateDownloading:   stateDownloadingHandler,
		backupApi.ArangoBackupStateDownloadError: stateDownloadErrorHandler,
		backupApi.ArangoBackupStateReady:         stateReadyHandler,
		backupApi.ArangoBackupStateDeleted:       stateDeletedHandler,
		backupApi.ArangoBackupStateFailed:        stateFailedHandler,
		backupApi.ArangoBackupStateUnavailable:   stateUnavailableHandler,
	}
)
