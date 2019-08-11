//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package v1alpha

import (
	"github.com/arangodb/kube-arangodb/pkg/backup/state"
)

const (
	ArangoBackupStateNone        state.State = ""
	ArangoBackupStatePending     state.State = "Pending"
	ArangoBackupStateScheduled   state.State = "Scheduled"
	ArangoBackupStateDownload    state.State = "Download"
	ArangoBackupStateDownloading state.State = "Downloading"
	ArangoBackupStateCreate      state.State = "Create"
	ArangoBackupStateUpload      state.State = "Upload"
	ArangoBackupStateUploading   state.State = "Uploading"
	ArangoBackupStateReady       state.State = "Ready"
	ArangoBackupStateDeleted     state.State = "Deleted"
	ArangoBackupStateFailed      state.State = "Failed"
)

var ArangoBackupStateMap = state.Map{
	ArangoBackupStateNone:        {ArangoBackupStatePending},
	ArangoBackupStatePending:     {ArangoBackupStateScheduled, ArangoBackupStateFailed},
	ArangoBackupStateScheduled:   {ArangoBackupStateDownload, ArangoBackupStateCreate, ArangoBackupStateFailed},
	ArangoBackupStateDownload:    {ArangoBackupStateDownloading, ArangoBackupStateFailed},
	ArangoBackupStateDownloading: {ArangoBackupStateReady, ArangoBackupStateFailed},
	ArangoBackupStateCreate:      {ArangoBackupStateReady, ArangoBackupStateDeleted, ArangoBackupStateUpload, ArangoBackupStateFailed},
	ArangoBackupStateUpload:      {ArangoBackupStateUploading, ArangoBackupStateFailed},
	ArangoBackupStateUploading:   {ArangoBackupStateReady, ArangoBackupStateFailed},
	ArangoBackupStateReady:       {ArangoBackupStateDeleted, ArangoBackupStateFailed},
	ArangoBackupStateDeleted:     {},
	ArangoBackupStateFailed:      {ArangoBackupStatePending},
}

type ArangoBackupState struct {
	// State holds the current high level state of the backup
	State state.State `json:"state"`

	// Message for the state this object is in.
	Message string `json:"Message,omitempty"`

	// Progress for the operation
	Progress *ArangoBackupProgress `json:"Message,omitempty"`
}

type ArangoBackupProgress struct {
	JobID    string `json:"jobID"`
	Progress string `json:"progress"`
}
