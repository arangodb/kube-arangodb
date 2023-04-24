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
	"strings"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createRotateOrUpgradePlan
func (r *Reconciler) createRebuildOutSyncedPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	if !features.RebuildOutSyncedShards().Enabled() {
		// RebuildOutSyncedShards feature is not enabled
		return nil
	}

	// TODO: use that to check if we need to rebuild shards
	r.log.Info("Rebuilding out-synced shards timeout", globals.GetGlobalTimeouts().ShardRebuild().Get())

	agencyState, ok := context.GetAgencyCache()
	if !ok {
		// Unable to get agency state, do not restart
		r.log.Error("Unable to get agency state")
		return nil
	}

	for _, m := range status.Members.AsList() {
		if m.Group == api.ServerGroupDBServers {
			// use * for global
			notInSyncShards := agency.GetDBServerShardsNotInSync(agencyState, agency.Server(m.Member.ID))

			if s := len(notInSyncShards); s > 0 {
				var shardsID []string
				for _, shard := range notInSyncShards {
					shardsID = append(shardsID, shard.Shard)
				}

				m.Member.Conditions.Update(api.ConditionTypeOutSyncedShards, true, "Member has out-synced shard(s)", strings.Join(shardsID, ", "))
				plan = append(plan, actions.NewAction(api.ActionTypeRebuildOutSyncedShards, api.ServerGroupDBServers, m.Member))
			}
		}
	}
	return plan
}
