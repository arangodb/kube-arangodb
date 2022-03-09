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

type CollectionIterator func(db, col string, planCollection *StatePlanCollection, shard string, plan ShardServers, current ShardServers)

func (s State) IterateOverCollections(i CollectionIterator) {
	for db, collections := range s.Plan.Collections {
		for collection, details := range collections {
			for shard, shardDetails := range details.Shards {
				currShard := s.Current.Collections[db][collection][shard]

				i(db, collection, &details, shard, shardDetails, currShard.Servers)
			}
		}
	}
}

func (s State) IsDBServerInSync(serverID string) bool {
	isInSync := true
	s.IterateOverCollections(func(db, col string, planCollection *StatePlanCollection, shard string, plan ShardServers, current ShardServers) {
		if !isInSync {
			return
		}
		serverIsNotInSync := plan.Contains(serverID) && !current.Contains(serverID)
		if serverIsNotInSync {
			isInSync = false
			return
		}
	})

	return isInSync
}

func (s State) IsDBServerReadyToRestart(serverID string) bool {
	readyToRestart := true

	s.IterateOverCollections(func(db, col string, planCollection *StatePlanCollection, shard string, plan ShardServers, current ShardServers) {
		if !readyToRestart {
			return
		}

		wc := planCollection.GetWriteConcern(0)
		rf := planCollection.GetReplicationFactor(0)

		if wc >= len(plan) && len(plan) != 0 {
			wc = len(plan) - 1
		}

		// Allow one server to be restarted if replication factor == write concern
		if rf > 0 && wc == rf {
			wc -= 1
		}

		serverIsNotInSync := plan.Contains(serverID) && !current.Contains(serverID)
		if len(current) >= wc && serverIsNotInSync {
			// Current shard is not in sync, but it does not matter - we have enough replicas in sync
			// Restart of this DBServer won't affect WC
			return
		}

		if len(current) < wc {
			// We have less in-sync servers than required for write concern
			readyToRestart = false
			return
		}
	})

	return readyToRestart
}
