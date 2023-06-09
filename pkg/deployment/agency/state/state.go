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

type Root struct {
	Arango   State `json:"arango"`
	ArangoDB DB    `json:"arangodb,omitempty"`
}

type DumpState struct {
	Agency Root `json:"agency"`
}

type DB struct {
	ArangoSync ArangoSyncLazy `json:"arangosync,omitempty"`
}

type State struct {
	Supervision Supervision `json:"Supervision"`
	Plan        Plan        `json:"Plan"`
	Current     Current     `json:"Current"`
	Target      Target      `json:"Target"`
}

type Current struct {
	MaintenanceServers CurrentMaintenanceServers `json:"MaintenanceServers,omitempty"`
	Collections        CurrentCollections        `json:"Collections"`
}

type Plan struct {
	Collections PlanCollections `json:"Collections"`
	Databases   PlanDatabases   `json:"Databases,omitempty"`
}

type Supervision struct {
	Maintenance Timestamp `json:"Maintenance,omitempty"`
}

func (s State) CountShards() int {
	count := 0

	for _, collections := range s.Plan.Collections {
		count += collections.CountShards()
	}

	return count
}

// ShardsByDBServers returns a map of DBServers and the amount of shards they have
func (s State) ShardsByDBServers() map[Server]int {
	result := make(map[Server]int)

	for _, collections := range s.Current.Collections {
		for _, shards := range collections {
			for _, shard := range shards {
				for _, server := range shard.Servers {
					result[server]++
				}
			}
		}
	}

	return result
}

// GetDBServerWithLowestShards returns the DBServer with the lowest amount of shards
func (s State) GetDBServerWithLowestShards() Server {
	var resultServer Server = ""
	var resultShards int

	for server, shards := range s.ShardsByDBServers() {
		// init first server as result
		if resultServer == "" {
			resultServer = server
			resultShards = shards
		} else if shards < resultShards {
			resultServer = server
			resultShards = shards
		}
	}
	return resultServer
}

type ShardDetails struct {
	ShardID    string
	Database   string
	Collection string
	Servers    Servers
}

// GetShardDetailsByID returns the ShardDetails for a given ShardID. If the ShardID is not found, the second return value is false
func (s State) GetShardDetailsByID(id string) (ShardDetails, bool) {
	// check first in Plan
	for dbName, db := range s.Plan.Collections {
		for colName, col := range db {
			for sName, servers := range col.Shards {
				if sName == id {
					return ShardDetails{
						ShardID:    sName,
						Database:   dbName,
						Collection: colName,
						Servers:    servers,
					}, true
				}
			}
		}
	}

	// check in Current
	for dbName, db := range s.Current.Collections {
		for colName, col := range db {
			for sName, shard := range col {
				if sName == id {
					return ShardDetails{
						ShardID:    sName,
						Database:   dbName,
						Collection: colName,
						Servers:    shard.Servers,
					}, true
				}
			}
		}
	}

	return ShardDetails{}, false
}

type ShardStatus struct {
	IsSynced bool
}

func (s State) GetShardsStatus() map[string]bool {
	q := map[string]bool{}

	for dName, d := range s.Plan.Collections {
		for cName, c := range d {
			for sName, servers := range c.Shards {
				q[sName] = s.IsShardInSync(dName, cName, sName, servers)
			}
		}
	}

	return q
}

func (s State) IsShardInSync(db, col, shard string, servers Servers) bool {
	dCurrent, ok := s.Current.Collections[db]
	if !ok {
		return false
	}

	cCurrent, ok := dCurrent[col]
	if !ok {
		return false
	}

	sCurrent, ok := cCurrent[shard]
	if !ok {
		return false
	}

	return sCurrent.Servers.InSync(servers)
}

// PlanServers returns all servers which are part of the plan
func (s State) PlanServers() Servers {
	q := map[Server]bool{}

	for _, db := range s.Plan.Collections {
		for _, col := range db {
			for _, shards := range col.Shards {
				for _, shard := range shards {
					q[shard] = true
				}
			}
		}
	}

	r := make([]Server, 0, len(q))

	for k := range q {
		r = append(r, k)
	}

	return r
}

// PlanLeaderServers returns all servers which are part of the plan as a leader
func (s State) PlanLeaderServers() Servers {
	q := map[Server]bool{}

	for _, db := range s.Plan.Collections {
		for _, col := range db {
			for _, shards := range col.Shards {
				if len(shards) == 0 {
					continue
				}
				q[shards[0]] = true
			}
		}
	}

	r := make([]Server, 0, len(q))

	for k := range q {
		r = append(r, k)
	}

	return r
}

type CollectionShardDetails []CollectionShardDetail

type CollectionShardDetail struct {
	Database   string
	Collection string
	Shard      string
}

type ShardFilter func(s State, db, col, shard string) bool

func NegateFilter(in ShardFilter) ShardFilter {
	return func(s State, db, col, shard string) bool {
		return !in(s, db, col, shard)
	}
}

func (s State) Filter(f ShardFilter) CollectionShardDetails {
	shards := make(CollectionShardDetails, s.CountShards())
	size := 0

	for db, collections := range s.Plan.Collections {
		for collection, details := range collections {
			for shard := range details.Shards {
				if f(s, db, collection, shard) {
					shards[size] = CollectionShardDetail{
						Database:   db,
						Collection: collection,
						Shard:      shard,
					}
					size++
				}
			}
		}
	}

	if size == 0 {
		return nil
	}

	return shards[0:size]
}

func GetDBServerBlockingRestartShards(s State, serverID Server) CollectionShardDetails {
	return s.Filter(FilterDBServerShardRestart(serverID))
}

func FilterDBServerShardRestart(serverID Server) ShardFilter {
	return NegateFilter(func(s State, db, col, shard string) bool {
		// Filter all shards which are not blocking restart of server
		plan := s.Plan.Collections[db][col]
		planShard := plan.Shards[shard]

		if !planShard.Contains(serverID) {
			// This DBServer is not even in plan, restart possible
			return true
		}

		current := s.Current.Collections[db][col][shard]
		currentShard := current.Servers.Join(planShard)

		serverInSync := currentShard.Contains(serverID)

		if len(planShard) == 1 && serverInSync {
			// The requested server is the only one in the plan, restart possible
			return true
		}

		wc := plan.GetWriteConcern(1)
		var rf int

		if plan.ReplicationFactor.IsUnknown() {
			// We are on unknown
			rf = len(currentShard)
		} else if plan.ReplicationFactor.IsSatellite() {
			// We are on satellite
			rf = len(s.PlanServers())
		} else {
			rf = int(plan.GetReplicationFactor(shard))
		}

		// If WriteConcern equals replicationFactor then downtime is always there
		if wc >= rf {
			wc = rf - 1
		}

		if len(currentShard) >= wc && !serverInSync {
			// Current shard is not in sync, but it does not matter - we have enough replicas in sync
			// Restart of this DBServer won't affect WC
			return true
		}

		if len(currentShard) > wc {
			// We are in plan, but restart is possible
			return true
		}

		// If we restart this server, write concern won't be satisfied
		return false
	})
}

func GetDBServerShardsNotInSync(s State, serverID Server) CollectionShardDetails {
	return s.Filter(FilterDBServerShardsNotInSync(serverID))
}

func FilterDBServerShardsNotInSync(serverID Server) ShardFilter {
	return NegateFilter(func(s State, db, col, shard string) bool {
		planShard := s.Plan.Collections[db][col].Shards[shard]

		if serverID != "*" && !planShard.Contains(serverID) {
			return true
		}

		currentShard := s.Current.Collections[db][col][shard]

		if len(planShard) != len(currentShard.Servers) {
			return false
		}

		for _, s := range planShard {
			if !currentShard.Servers.Contains(s) {
				return false
			}
		}

		return true
	})
}
