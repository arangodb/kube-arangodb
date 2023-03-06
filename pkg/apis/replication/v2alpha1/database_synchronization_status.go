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

package v2alpha1

// DatabaseSynchronizationStatus contains the synchronization status of replication for database
type DatabaseSynchronizationStatus struct {
	// Deprecated
	// ShardsTotal shows how many shards are expected to be in-sync
	ShardsTotal int `json:"shards-total"`
	// Deprecated
	// ShardsInSync shows how many shards are already in-sync
	ShardsInSync int `json:"shards-in-sync"`
	// Errors contains a list of errors if there were unexpected errors during synchronization
	Errors []DatabaseSynchronizationError `json:"errors,omitempty"`
}

// DatabaseSynchronizationError contains the error message for specific shard in collection
type DatabaseSynchronizationError struct {
	Collection string `json:"collection,omitempty"`
	Shard      string `json:"shard,omitempty"`
	Message    string `json:"message"`
}
