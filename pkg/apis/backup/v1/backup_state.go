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

package v1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/handlers/backup/state"
)

const (
	ArangoBackupStateNone          state.State = ""
	ArangoBackupStatePending       state.State = "Pending"
	ArangoBackupStateScheduled     state.State = "Scheduled"
	ArangoBackupStateDownload      state.State = "Download"
	ArangoBackupStateDownloadError state.State = "DownloadError"
	ArangoBackupStateDownloading   state.State = "Downloading"
	ArangoBackupStateCreate        state.State = "Create"
	ArangoBackupStateCreateError   state.State = "CreateError"
	ArangoBackupStateUpload        state.State = "Upload"
	ArangoBackupStateUploading     state.State = "Uploading"
	ArangoBackupStateUploadError   state.State = "UploadError"
	ArangoBackupStateReady         state.State = "Ready"
	ArangoBackupStateDeleted       state.State = "Deleted"
	ArangoBackupStateFailed        state.State = "Failed"
	ArangoBackupStateUnavailable   state.State = "Unavailable"
)

var ArangoBackupStateMap = state.Map{
	ArangoBackupStateNone:          {ArangoBackupStatePending, ArangoBackupStateFailed},
	ArangoBackupStatePending:       {ArangoBackupStateScheduled, ArangoBackupStateFailed},
	ArangoBackupStateScheduled:     {ArangoBackupStateDownload, ArangoBackupStateCreate, ArangoBackupStateFailed},
	ArangoBackupStateDownload:      {ArangoBackupStateDownloading, ArangoBackupStateFailed, ArangoBackupStateDownloadError},
	ArangoBackupStateDownloading:   {ArangoBackupStateReady, ArangoBackupStateFailed, ArangoBackupStateDownloadError},
	ArangoBackupStateDownloadError: {ArangoBackupStatePending, ArangoBackupStateFailed},
	ArangoBackupStateCreate:        {ArangoBackupStateReady, ArangoBackupStateFailed, ArangoBackupStateCreateError},
	ArangoBackupStateCreateError:   {ArangoBackupStateFailed, ArangoBackupStateCreate},
	ArangoBackupStateUpload:        {ArangoBackupStateUploading, ArangoBackupStateFailed, ArangoBackupStateDeleted, ArangoBackupStateUploadError},
	ArangoBackupStateUploading:     {ArangoBackupStateReady, ArangoBackupStateFailed, ArangoBackupStateUploadError},
	ArangoBackupStateUploadError:   {ArangoBackupStateFailed, ArangoBackupStateReady},
	ArangoBackupStateReady:         {ArangoBackupStateDeleted, ArangoBackupStateFailed, ArangoBackupStateUpload, ArangoBackupStateUnavailable},
	ArangoBackupStateDeleted:       {ArangoBackupStateFailed, ArangoBackupStateReady},
	ArangoBackupStateFailed:        {ArangoBackupStatePending},
	ArangoBackupStateUnavailable:   {ArangoBackupStateReady, ArangoBackupStateDeleted, ArangoBackupStateFailed},
}

type ArangoBackupState struct {
	// State holds the current high level state of the backup
	// +doc/enum: Pending|state in which Custom Resource is queued. If Backup is possible changed to "Scheduled"
	// +doc/enum: Scheduled|state which will start create/download process
	// +doc/enum: Download|state in which download request will be created on ArangoDB
	// +doc/enum: DownloadError|state when download failed
	// +doc/enum: Downloading|state for downloading progress
	// +doc/enum: Create|state for creation, field available set to true
	// +doc/enum: Upload|state in which upload request will be created on ArangoDB
	// +doc/enum: Uploading|state for uploading progress
	// +doc/enum: UploadError|state when uploading failed
	// +doc/enum: Ready|state when Backup is finished
	// +doc/enum: Deleted|state when Backup was once in ready, but has been deleted
	// +doc/enum: Failed|state for failure
	// +doc/enum: Unavailable|state when Backup is not available on the ArangoDB. It can happen in case of upgrades, node restarts etc.
	State state.State `json:"state"`

	// Time is the time in UTC when state of the ArangoBackup Custom Resource changed.
	Time meta.Time `json:"time"`

	// Message for the state this object is in.
	Message string `json:"message,omitempty"`

	// Progress info of the uploading and downloading process
	Progress *ArangoBackupProgress `json:"progress,omitempty"`
}

func (a *ArangoBackupState) Equal(b *ArangoBackupState) bool {
	if a == b {
		return true
	}

	if a == nil && b != nil || a != nil && b == nil {
		return false
	}

	return a.State == b.State &&
		a.Time.Equal(&b.Time) &&
		a.Message == b.Message &&
		a.Progress.Equal(b.Progress)
}

type ArangoBackupProgress struct {
	// JobID ArangoDB job ID for uploading or downloading
	JobID string `json:"jobID"`
	// Progress ArangoDB job progress in percents
	// +doc/example: 90%
	Progress string `json:"progress"`
}

func (a *ArangoBackupProgress) Equal(b *ArangoBackupProgress) bool {
	if a == b {
		return true
	}

	if a == nil && b != nil || a != nil && b == nil {
		return false
	}

	return a.JobID == b.JobID &&
		a.Progress == b.Progress
}
