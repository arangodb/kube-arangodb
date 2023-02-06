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

// SynchronizationStatus contains the synchronization status of replication for all databases
type SynchronizationStatus struct {
	// Databases holds the synchronization status for each database.
	Databases map[string]DatabaseSynchronizationStatus `json:"databases,omitempty"`
	// Progress the value in percents showing the progress of synchronization
	Progress float32 `json:"progress,omitempty"`
	// AllInSync is true if target cluster is fully in-sync with source cluster
	AllInSync bool `json:"allInSync,omitempty"`
	// Error contains an error description if there is an error preventing synchronization
	Error string `json:"error,omitempty"`
}
