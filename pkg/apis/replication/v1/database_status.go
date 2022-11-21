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

// DatabaseStatus contains the status of a single database.
// Deprecated
type DatabaseStatus struct {
	// Name of the database
	Name string `json:"name"`
	// Collections holds the replication status of each collection in the database.
	// List is ordered by name of the collection.
	Collections []CollectionStatus `json:"collections,omitempty"`
}
