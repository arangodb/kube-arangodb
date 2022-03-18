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

func (a StatePlanDBCollections) CountShards() int {
	count := 0

	for _, d := range a {
		count += len(d.Shards)
	}

	return count
}

type StatePlanCollection struct {
	Name   *string `json:"name"`
	Shards Shards  `json:"shards"`
	// deprecated
	// MinReplicationFactor is deprecated, but we have to support it for backward compatibility
	MinReplicationFactor *int `json:"minReplicationFactor,omitempty"`
	WriteConcern         *int `json:"writeConcern,omitempty"`
	ReplicationFactor    *int `json:"replicationFactor,omitempty"`
}

func (a *StatePlanCollection) GetReplicationFactor(shard string) int {
	if a == nil {
		return 0
	}

	l := len(a.Shards[shard])

	if z := a.ReplicationFactor; z == nil {
		return l
	} else {
		if v := *z; v > l {
			return v
		} else {
			return l
		}
	}
}

func (a *StatePlanCollection) GetWriteConcern(def int) int {
	if p := a.GetWriteConcernP(); p != nil {
		return *p
	}

	return def
}

func (a *StatePlanCollection) GetWriteConcernP() *int {
	if a == nil {
		return nil
	}

	if a.WriteConcern == nil {
		return a.MinReplicationFactor
	}

	return a.WriteConcern
}

func (a StatePlanCollection) GetName(d string) string {
	if a.Name == nil {
		return d
	}

	return *a.Name
}

func (a *StatePlanCollection) IsDBServerInShards(name string) bool {
	if a == nil {
		return false
	}

	for _, planShards := range a.Shards {
		if planShards.Contains(name) {
			return true
		}
	}
	return false
}
