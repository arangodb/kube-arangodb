//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package agency

type ArangoPlanDatabases map[string]ArangoPlanCollections

func (a ArangoPlanDatabases) IsDBServerInDatabases(name string) bool {
	for _, collections := range a {
		if collections.IsDBServerInCollections(name) {
			return true
		}
	}
	return false
}

type ArangoPlanCollections map[string]ArangoPlanCollection

func (a ArangoPlanCollections) IsDBServerInCollections(name string) bool {
	for _, collection := range a {
		if collection.IsDBServerInShards(name) {
			return true
		}
	}
	return false
}

type ArangoPlanCollection struct {
	Shards ArangoPlanShard `json:"shards"`
}

func (a ArangoPlanCollection) IsDBServerInShards(name string) bool {
	for _, dbservers := range a.Shards {
		for _, dbserver := range dbservers {
			if dbserver == name {
				return true
			}
		}
	}
	return false
}

type ArangoPlanShard map[string][]string