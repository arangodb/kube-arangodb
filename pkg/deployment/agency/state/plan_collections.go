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

package state

// PlanCollections is a map of database name to collections
type PlanCollections map[string]PlanDBCollections

func (a PlanCollections) IsDBServerPresent(name Server) bool {
	for _, collections := range a {
		if collections.IsDBServerInCollections(name) {
			return true
		}
	}
	return false
}

func (a PlanCollections) IsDBServerLeader(name Server) bool {
	for _, collections := range a {
		if collections.IsDBServerLeaderInCollections(name) {
			return true
		}
	}
	return false
}

// PlanDBCollections is a map of collection name to collection details
type PlanDBCollections map[string]PlanCollection

func (a PlanDBCollections) IsDBServerInCollections(name Server) bool {
	for _, collection := range a {
		if collection.IsDBServerInShards(name) {
			return true
		}
	}
	return false
}

func (a PlanDBCollections) IsDBServerLeaderInCollections(name Server) bool {
	for _, collection := range a {
		if collection.IsDBServerLeader(name) {
			return true
		}
	}
	return false
}

func (a PlanDBCollections) CountShards() int {
	count := 0

	for _, d := range a {
		count += len(d.Shards)
	}

	return count
}

type PlanCollection struct {
	Name   *string `json:"name"`
	Shards Shards  `json:"shards"`
	// deprecated
	// MinReplicationFactor is deprecated, but we have to support it for backward compatibility
	MinReplicationFactor *int               `json:"minReplicationFactor,omitempty"`
	WriteConcern         *int               `json:"writeConcern,omitempty"`
	ReplicationFactor    *ReplicationFactor `json:"replicationFactor,omitempty"`
	DistributeShardsLike *string            `json:"distributeShardsLike,omitempty"`
}

func (a *PlanCollection) GetReplicationFactor(shard string) ReplicationFactor {
	if a == nil {
		return 0
	}

	l := ReplicationFactor(len(a.Shards[shard]))

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

func (a *PlanCollection) GetWriteConcern(def int) int {
	if p := a.GetWriteConcernP(); p != nil {
		return *p
	}

	return def
}

func (a *PlanCollection) GetWriteConcernP() *int {
	if a == nil {
		return nil
	}

	if a.WriteConcern == nil {
		return a.MinReplicationFactor
	}

	return a.WriteConcern
}

func (a PlanCollection) GetName(d string) string {
	if a.Name == nil {
		return d
	}

	return *a.Name
}

func (a *PlanCollection) IsDBServerInShards(name Server) bool {
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

func (a *PlanCollection) IsDBServerLeader(name Server) bool {
	if a == nil {
		return false
	}

	for _, servers := range a.Shards {
		if len(servers) == 0 {
			continue
		}
		if servers[0] == name {
			return true
		}
	}
	return false
}
