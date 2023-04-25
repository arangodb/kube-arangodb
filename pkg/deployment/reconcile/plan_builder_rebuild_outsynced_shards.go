//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var lastTriggeredRebuildOutSyncedShards time.Time

// createRotateOrUpgradePlan
func (r *Reconciler) createRebuildOutSyncedPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	if !features.RebuildOutSyncedShards().Enabled() {
		// RebuildOutSyncedShards feature is not enabled
		return nil
	}

	// to prevent from rebuilding out-synced shards again and again we will trigger rebuild only once per T minutes
	if time.Since(lastTriggeredRebuildOutSyncedShards) < globals.GetGlobalTimeouts().ShardRebuildRetry().Get() {
		// we already triggered rebuild out-synced shards recently
		return nil
	}

	shardStatus, ok := context.ShardsInSyncMap()
	if !ok {
		r.log.Error("Unable to get shards status")
		return nil
	}
	agencyState, ok := context.GetAgencyCache()
	if !ok {
		r.log.Error("Unable to get agency state")
		return nil
	}

	members := map[string]api.MemberStatus{}
	for _, m := range status.Members.AsListInGroup(api.ServerGroupDBServers) {
		members[m.Member.ID] = m.Member
	}

	// Get shards which are out-synced for more than defined timeout
	outSyncedShardsIDs := shardStatus.NotInSyncSince(globals.GetGlobalTimeouts().ShardRebuild().Get())

	if len(outSyncedShardsIDs) > 0 {
		// Create plan for out-synced shards
		for _, shardID := range outSyncedShardsIDs {
			shard, exist := agencyState.GetShardDetailsByID(shardID)
			if !exist {
				r.log.Error("Shard servers not found", shardID, shard.Database)
				continue
			}

			for _, server := range shard.Servers {
				member, ok := members[string(server)]
				if !ok {
					r.log.Error("Member not found - we can not fix out-synced shard!", server)
				} else {
					r.log.Info("Shard is out-synced and its Tree will be rebuild", shardID, shard.Database, shard.Collection, member.ID)

					action := actions.NewAction(api.ActionTypeRebuildOutSyncedShards, api.ServerGroupDBServers, member).
						AddParam("shardID", shardID).
						AddParam("database", shard.Database)

					plan = append(plan, action)
				}
			}
		}

		// save time when we triggered rebuild out-synced shards last time
		lastTriggeredRebuildOutSyncedShards = time.Now()
	}
	return plan
}
