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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArangoBackupStatus contains the status part of
// an ArangoBackup.
type ArangoBackupStatus struct {
	ArangoBackupState `json:",inline"`
	Backup            *ArangoBackupDetails `json:"backup,omitempty"`
	Available         bool                 `json:"available"`
}

type ArangoBackupDetails struct {
	ID                      string    `json:"id"`
	Version                 string    `json:"version"`
	PotentiallyInconsistent *bool     `json:"potentiallyInconsistent,omitempty"`
	Uploaded                *bool     `json:"uploaded,omitempty"`
	Downloaded              *bool     `json:"downloaded,omitempty"`
	Imported                *bool     `json:"imported,omitempty"`
	CreationTimestamp       meta.Time `json:"createdAt"`
}
