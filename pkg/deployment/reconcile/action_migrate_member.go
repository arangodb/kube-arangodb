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

package reconcile

import (
	"context"
	"fmt"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

const (
	actionMigrateMemberSourceKey string = "source-member-id"
	actionMigrateMemberBatchSize int    = 64
)

// newMigrateMemberAction creates a new Action that implements the given
// planned MigrateMember action.
func newMigrateMemberAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionMigrateMember{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionMigrateMember implements an MigrateMemberAction.
type actionMigrateMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.git
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionMigrateMember) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	return ready, err
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionMigrateMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	if !features.MemberReplaceMigration().Enabled() {
		// Feature disabled
		return true, false, nil
	}

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Warn("Unable to find MemberID %s", a.action.MemberID)
		return true, false, nil
	}

	sourceMemberID, ok := a.action.GetParam(actionMigrateMemberSourceKey)
	if !ok {
		a.log.Warn("Unable to find action's param %s", actionMigrateMemberSourceKey)
		return true, false, nil
	}

	sourceMember, ok := a.actionCtx.GetMemberStatusByID(sourceMemberID)
	if !ok {
		a.log.Warn("Unable to find member %s", sourceMemberID)
		return true, false, nil
	}

	if a.action.Group != api.ServerGroupDBServers {
		// Proceed only on DBServers
		a.log.Warn("Member %s is not DBServer", a.action.MemberID)
		return true, false, nil
	}

	cache, ok := a.actionCtx.GetAgencyCache()
	if !ok {
		a.log.Debug("AgencyCache is not ready")
		return false, false, nil
	}

	if !cache.Plan.DBServers.Exists(state.Server(sourceMember.ID)) {
		a.log.JSON("databases", cache.Plan.DBServers).Str("id", sourceMember.ID).Debug("Source DBServer not yet present")
		return true, false, nil
	}

	if !cache.Plan.DBServers.Exists(state.Server(m.ID)) {
		a.log.JSON("databases", cache.Plan.DBServers).Str("id", m.ID).Debug("Destination DBServer not yet present")
		return false, false, nil
	}

	stats := cache.PlanServerUsage(state.Server(sourceMember.ID))

	if stats.Count() == 0 {
		a.log.Debug("DBServer not in use anymore")
		// Server not in use anymore
		return true, false, nil
	}

	c, err := a.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		a.log.Err(err).Error("Unable to get client")
		return false, false, nil
	}

	clusterClient := client.NewClient(c.Connection(), a.log)

	resp, err := clusterClient.RebalanceGet(ctx)
	if err != nil {
		a.log.Err(err).Error("Unable to get rebalance status")
		return false, false, nil
	}

	if resp.Result.PendingMoveShards != 0 || resp.Result.TodoMoveShards != 0 {
		a.actionCtx.SetProgress(fmt.Sprintf("Currently rebalance in progress, Shards: %d. Jobs: %d", stats.Count(), resp.Result.PendingMoveShards+resp.Result.TodoMoveShards))
		return false, false, nil
	}

	var moves = make(client.RebalanceExecuteRequestMoves, 0, actionMigrateMemberBatchSize)

	for db, collections := range cache.Plan.Collections {
		if len(moves) >= actionMigrateMemberBatchSize {
			break
		}
		for collection, details := range collections {
			if len(moves) >= actionMigrateMemberBatchSize {
				break
			}

			if details.DistributeShardsLike != nil {
				continue
			}

			if details.ReplicationFactor.IsSatellite() {
				continue
			}

			for shard, servers := range details.Shards {
				if len(moves) >= actionMigrateMemberBatchSize {
					break
				}

				if len(servers) == 0 {
					continue
				}

				if !servers.Contains(state.Server(sourceMember.ID)) {
					continue
				}

				if servers.Contains(state.Server(m.ID)) {
					continue
				}

				a.log.Str("db", db).
					Str("collection", collection).
					Str("shard", shard).
					Str("from", sourceMember.ID).
					Str("to", m.ID).
					Bool("leader", string(servers[0]) == sourceMember.ID).Debug("Migrating shard")

				moves = append(moves, client.RebalanceExecuteRequestMove{
					Database:   db,
					Collection: collection,
					Shard:      shard,
					From:       sourceMember.ID,
					To:         m.ID,
					IsLeader:   string(servers[0]) == sourceMember.ID,
				})
			}
		}
	}

	if len(moves) > 0 {
		if err := clusterClient.RebalanceExecuteMoves(ctx, moves...); err != nil {
			a.log.Err(err).Error("Unable to execute rebalance status")
			return false, false, nil
		}

		return false, false, nil
	}

	// Cleanout completed
	return true, false, nil
}
