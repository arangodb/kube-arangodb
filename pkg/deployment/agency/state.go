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

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/agency"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func loadState(ctx context.Context, client agency.Agency) (State, error) {
	conn := client.Connection()

	req, err := client.Connection().NewRequest(http.MethodPost, "/_api/agency/read")
	if err != nil {
		return State{}, err
	}

	var data []byte

	req, err = req.SetBody(GetAgencyReadRequest(GetAgencyReadKey(GetAgencyKey(ArangoKey, SupervisionKey, SupervisionMaintenanceKey), GetAgencyKey(ArangoKey, PlanKey, PlanCollectionsKey), GetAgencyKey(ArangoKey, CurrentKey, PlanCollectionsKey))))
	if err != nil {
		return State{}, err
	}

	resp, err := conn.Do(driver.WithRawResponse(ctx, &data), req)
	if err != nil {
		return State{}, err
	}

	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return State{}, err
	}

	var c StateRoots

	if err := json.Unmarshal(data, &c); err != nil {
		return State{}, err
	}

	if len(c) != 1 {
		return State{}, errors.Newf("Invalid response size")
	}

	state := c[0].Arango

	if _, ok := state.Current.Collections["_system"]; !ok {
		return State{}, errors.Newf("Unable to find system database (invalid data)")
	}

	if _, ok := state.Plan.Collections["_system"]; !ok {
		return State{}, errors.Newf("Unable to find system database (invalid data)")
	}

	return state, nil
}

type StateRoots []StateRoot

type StateRoot struct {
	Arango State `json:"arango"`
}

type DumpState struct {
	Agency StateRoot `json:"agency"`
}

type State struct {
	Supervision StateSupervision `json:"Supervision"`
	Plan        StatePlan        `json:"Plan"`
	Current     StateCurrent     `json:"Current"`
}

type StateCurrent struct {
	Collections StateCurrentCollections `json:"Collections"`
}

type StatePlan struct {
	Collections StatePlanCollections `json:"Collections"`
}

type StateSupervision struct {
	Maintenance StateExists `json:"Maintenance,omitempty"`
}

type StateExists bool

func (d *StateExists) Exists() bool {
	if d == nil {
		return false
	}

	return bool(*d)
}

func (d *StateExists) UnmarshalJSON(bytes []byte) error {
	*d = bytes != nil
	return nil
}

func (s State) CountShards() int {
	count := 0

	for _, collections := range s.Plan.Collections {
		count += collections.CountShards()
	}

	return count
}

func (s State) PlanServers() []string {
	q := map[string]bool{}

	for _, db := range s.Plan.Collections {
		for _, col := range db {
			for _, shards := range col.Shards {
				for _, shard := range shards {
					q[shard] = true
				}
			}
		}
	}

	r := make([]string, 0, len(q))

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

func GetDBServerBlockingRestartShards(s State, serverID string) CollectionShardDetails {
	return s.Filter(FilterDBServerShardRestart(serverID))
}

func FilterDBServerShardRestart(serverID string) StateShardFilter {
	return NegateFilter(func(s State, db, col, shard string) bool {
		// Filter all shards which are not blocking restart of server
		plan := s.Plan.Collections[db][col]
		planShard := plan.Shards[shard]

		if !planShard.Contains(serverID) {
			// This DBServer is not even in plan, restart possible
			return true
		}

		current := s.Current.Collections[db][col][shard]
		currentShard := current.Servers.FilterBy(planShard)

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

func GetDBServerShardsNotInSync(s State, serverID string) CollectionShardDetails {
	return s.Filter(FilterDBServerShardsNotInSync(serverID))
}

func FilterDBServerShardsNotInSync(serverID string) StateShardFilter {
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
