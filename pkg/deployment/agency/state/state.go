//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

// ServerKnown stores information about single ArangoDB server.
type ServerKnown struct {
	// RebootID is an incremental value which describes how many times server was restarted.
	RebootID int `json:"rebootId"`
}

type Current struct {
	MaintenanceDBServers CurrentMaintenanceDBServers `json:"MaintenanceDBServers,omitempty"`
	Collections          CurrentCollections          `json:"Collections"`

	// ServersKnown stores information about ArangoDB servers.
	ServersKnown ServerMap[ServerKnown] `json:"ServersKnown,omitempty"`
}

type Plan struct {
	Collections  PlanCollections   `json:"Collections"`
	Databases    PlanDatabases     `json:"Databases,omitempty"`
	DBServers    ServerMap[string] `json:"DBServers,omitempty"`
	Coordinators ServerMap[string] `json:"Coordinators,omitempty"`
}

type ShardCountDetails struct {
	Leader, Follower int
}

type ServerMap[T any] map[Server]T

func (s ServerMap[T]) Exists(server Server) bool {
	_, ok := s[server]
	return ok
}

func (s ShardCountDetails) Count() int {
	return s.Leader + s.Follower
}

func (s ShardCountDetails) Add(leader bool) ShardCountDetails {
	if leader {
		s.Leader += 1
	} else {
		s.Follower += 1
	}

	return s
}

func (s State) CountShards() int {
	count := 0

	for _, collections := range s.Plan.Collections {
		count += collections.CountShards()
	}

	return count
}

// ShardsByDBServers returns a map of DBServers and the amount of shards they have
func (s State) ShardsByDBServers() map[Server]ShardCountDetails {
	result := make(map[Server]ShardCountDetails)

	for _, collections := range s.Current.Collections {
		for _, shards := range collections {
			for _, shard := range shards {
				for id, server := range shard.Servers {
					result[server] = result[server].Add(id == 0)
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
			resultShards = shards.Count()
		} else if shards.Count() < resultShards {
			resultServer = server
			resultShards = shards.Count()
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

// PlanServerUsage returns number of the shards and replicas by a server
func (s State) PlanServerUsage(id Server) ShardCountDetails {
	var z ShardCountDetails
	for _, db := range s.Plan.Collections {
		for _, col := range db {
			for _, shards := range col.Shards {
				for i, shard := range shards {
					if shard == id {
						z = z.Add(i == 0)
					}
				}
			}
		}
	}

	return z
}

// PlanLeaderServersWithFailOver returns all servers which are part of the plan as a leader and can fail over
func (s State) PlanLeaderServersWithFailOver() Servers {
	q := map[Server]bool{}

	for _, db := range s.Plan.Collections {
		for _, col := range db {
			for _, shards := range col.Shards {
				if len(shards) <= 1 {
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
		if _, ok := s.Plan.Databases[db]; !ok {
			// DB Is missing, restart possible
			return true
		}

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

// GetCollectionDatabaseByID find Database name by Collection ID
func (s State) GetCollectionDatabaseByID(id string) (string, bool) {
	for db, cols := range s.Current.Collections {
		if _, ok := cols[id]; ok {
			return db, true
		}
	}

	return "", false
}

// GetRebootID returns reboot ID for a given server ID.
// returns false when a server ID does not exist in cache.
func (s State) GetRebootID(id Server) (int, bool) {
	if v, ok := s.Current.ServersKnown[id]; ok {
		return v.RebootID, true
	}

	return 0, false
}
