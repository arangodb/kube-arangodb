//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

package agency

type StatePlanCollections map[string]StatePlanDBCollections

func (a StatePlanCollections) IsDBServerInDatabases(name string) bool {
	for _, collections := range a {
		if collections.IsDBServerInCollections(name) {
			return true
		}
	}
	return false
}

type StatePlanDBCollections map[string]StatePlanCollection

func (a StatePlanDBCollections) IsDBServerInCollections(name string) bool {
	for _, collection := range a {
		if collection.IsDBServerInShards(name) {
			return true
		}
	}
	return false
}

type StatePlanCollection struct {
	Name   *string        `json:"name"`
	Shards StatePlanShard `json:"shards"`
}

func (a StatePlanCollection) GetName(d string) string {
	if a.Name == nil {
		return d
	}

	return *a.Name
}

func (a StatePlanCollection) IsDBServerInShards(name string) bool {
	for _, dbservers := range a.Shards {
		for _, dbserver := range dbservers {
			if dbserver == name {
				return true
			}
		}
	}
	return false
}

type StatePlanShard map[string][]string
