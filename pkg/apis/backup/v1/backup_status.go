//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

// ArangoBackupStatus contains the status part of
// an ArangoBackup.
type ArangoBackupStatus struct {
	ArangoBackupState `json:",inline"`
	Backup            *ArangoBackupDetails       `json:"backup,omitempty"`
	Available         bool                       `json:"available"`
	Backoff           *ArangoBackupStatusBackOff `json:"backoff,omitempty"`
}

func (a *ArangoBackupStatus) Equal(b *ArangoBackupStatus) bool {
	if a == b {
		return true
	}

	if a == nil && b != nil || a != nil && b == nil {
		return false
	}

	return a.ArangoBackupState.Equal(&b.ArangoBackupState) &&
		a.Backup.Equal(b.Backup) &&
		a.Available == b.Available
}

type ArangoBackupDetails struct {
	ID                      string          `json:"id"`
	Version                 string          `json:"version"`
	PotentiallyInconsistent *bool           `json:"potentiallyInconsistent,omitempty"`
	SizeInBytes             uint64          `json:"sizeInBytes,omitempty"`
	NumberOfDBServers       uint            `json:"numberOfDBServers,omitempty"`
	Uploaded                *bool           `json:"uploaded,omitempty"`
	Downloaded              *bool           `json:"downloaded,omitempty"`
	Imported                *bool           `json:"imported,omitempty"`
	CreationTimestamp       meta.Time       `json:"createdAt"`
	Keys                    shared.HashList `json:"keys,omitempty"`
}

func (a *ArangoBackupDetails) Equal(b *ArangoBackupDetails) bool {
	if a == b {
		return true
	}

	if a == nil && b != nil || a != nil && b == nil {
		return false
	}

	return a.ID == b.ID &&
		a.Version == b.Version &&
		a.SizeInBytes == b.SizeInBytes &&
		a.NumberOfDBServers == b.NumberOfDBServers &&
		a.CreationTimestamp.Equal(&b.CreationTimestamp) &&
		compareBoolPointer(a.PotentiallyInconsistent, b.PotentiallyInconsistent) &&
		compareBoolPointer(a.Uploaded, b.Uploaded) &&
		compareBoolPointer(a.Downloaded, b.Downloaded) &&
		compareBoolPointer(a.Imported, b.Imported) &&
		a.Keys.Equal(b.Keys)
}

func compareBoolPointer(a, b *bool) bool {
	if a == nil && b != nil || a != nil && b == nil {
		return false
	}

	if a == b {
		return true
	}

	if *a == *b {
		return true
	}

	return false
}
