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

package agency

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (c *cache) loadState(ctx context.Context, client agency.Agency) (StateRoot, error) {
	conn := client.Connection()

	req, err := client.Connection().NewRequest(http.MethodPost, "/_api/agency/read")
	if err != nil {
		return StateRoot{}, err
	}

	var data []byte

	readKeys := []string{
		GetAgencyKey(ArangoKey, SupervisionKey, SupervisionMaintenanceKey),
		GetAgencyKey(ArangoKey, PlanKey, PlanCollectionsKey),
		GetAgencyKey(ArangoKey, PlanKey, PlanDatabasesKey),
		GetAgencyKey(ArangoKey, CurrentKey, PlanCollectionsKey),
		GetAgencyKey(ArangoKey, CurrentKey, CurrentMaintenanceServers),
		GetAgencyKey(ArangoKey, TargetKey, TargetHotBackupKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetJobToDoKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetJobPendingKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetJobFailedKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetJobFinishedKey),
		GetAgencyKey(ArangoKey, TargetKey, TargetCleanedServersKey),
		GetAgencyKey(ArangoDBKey, ArangoSyncKey, ArangoSyncStateKey, ArangoSyncStateIncomingKey, ArangoSyncStateIncomingStateKey),
		GetAgencyKey(ArangoDBKey, ArangoSyncKey, ArangoSyncStateKey, ArangoSyncStateOutgoingKey, ArangoSyncStateOutgoingTargetsKey),
	}

	req, err = req.SetBody(GetAgencyReadRequest(GetAgencyReadKey(readKeys...)))
	if err != nil {
		return StateRoot{}, err
	}

	resp, err := conn.Do(driver.WithRawResponse(ctx, &data), req)
	if err != nil {
		return StateRoot{}, err
	}

	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return StateRoot{}, err
	}

	var r StateRoots

	if err := json.Unmarshal(data, &r); err != nil {
		return StateRoot{}, err
	}

	if len(r) != 1 {
		return StateRoot{}, errors.Newf("Invalid response size")
	}

	state := r[0]

	return state, nil
}

type StateRoots []StateRoot

type StateRoot struct {
	Arango   State   `json:"arango"`
	ArangoDB StateDB `json:"arangodb,omitempty"`
}

type DumpState struct {
	Agency StateRoot `json:"agency"`
}

type StateDB struct {
	ArangoSync ArangoSyncLazy `json:"arangosync,omitempty"`
}

type State struct {
	Supervision StateSupervision `json:"Supervision"`
	Plan        StatePlan        `json:"Plan"`
	Current     StateCurrent     `json:"Current"`
	Target      StateTarget      `json:"Target"`
}

type StateCurrent struct {
	MaintenanceServers StateCurrentMaintenanceServers `json:"MaintenanceServers,omitempty"`
	Collections        StateCurrentCollections        `json:"Collections"`
}

type StatePlan struct {
	Collections StatePlanCollections `json:"Collections"`
	Databases   PlanDatabases        `json:"Databases,omitempty"`
}

type StateSupervision struct {
	Maintenance StateTimestamp `json:"Maintenance,omitempty"`
}

func (s State) CountShards() int {
	count := 0

	for _, collections := range s.Plan.Collections {
		count += collections.CountShards()
	}

	return count
}

// ShardsByDbServers returns a map of DBServers and the amount of shards they have
func (s State) ShardsByDbServers() map[Server]int {
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

	for server, shards := range s.ShardsByDbServers() {
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

func (s State) GetShardsStatus() map[string]bool {
	q := map[string]bool{}

	for dName, d := range s.Plan.Collections {
		for cName, c := range d {
			for sName, shard := range c.Shards {
				q[sName] = s.IsShardInSync(dName, cName, sName, shard)
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

type CollectionShardDetails []CollectionShardDetail

type CollectionShardDetail struct {
	Database   string
	Collection string
	Shard      string
}

type StateShardFilter func(s State, db, col, shard string) bool

func NegateFilter(in StateShardFilter) StateShardFilter {
	return func(s State, db, col, shard string) bool {
		return !in(s, db, col, shard)
	}
}

func (s State) Filter(f StateShardFilter) CollectionShardDetails {
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

func FilterDBServerShardRestart(serverID Server) StateShardFilter {
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

func FilterDBServerShardsNotInSync(serverID Server) StateShardFilter {
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
