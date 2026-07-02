//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (r *Reconciler) createMemberMaintenanceManagementPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {

	if !features.Version310().Enabled() {
		return nil
	}

	var plan api.Plan
	for _, member := range status.Members.AsListInGroups(api.ServerGroupDBServers) {
		if member.Member.Conditions.IsTrue(api.ConditionTypeMemberMaintenanceMode) {
			plan = append(plan, actions.NewAction(api.ActionTypeDisableMemberMaintenance, member.Group, member.Member, "Disable maintenance due to missing plan"))
		}
	}
	return plan
}

// createHighMemberMaintenanceDisablePlan emits DisableMemberMaintenance for any DBServer that has
// member maintenance enabled, is no longer Ready and currently leads shards that can fail over to
// another replica. Runs unconditionally (via Apply, not ApplyIfEmpty) so it fires even while the
// normal plan is busy with recovery/restart actions.
func (r *Reconciler) createHighMemberMaintenanceDisablePlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {

	if !features.Version310().Enabled() {
		return nil
	}

	agencyState, ok := planCtx.GetAgencyCache()
	if !ok {
		// Without the agency state we cannot tell whether the member hosts leaders, so do nothing.
		return nil
	}

	// Only consider leaders on shards with a follower to fail over to. Single-replica (RF1) shards
	// have nowhere to fail over, so disabling maintenance for a member that only leads RF1 shards
	// would not restore availability - leave maintenance enabled in that case.
	leaders := agencyState.PlanLeaderServersWithFailOver()

	var plan api.Plan
	for _, member := range status.Members.AsListInGroups(api.ServerGroupDBServers) {
		readyCond, hasReady := member.Member.Conditions.Get(api.ConditionTypeReady)
		if !(member.Member.Conditions.IsTrue(api.ConditionTypeMemberMaintenanceMode) &&
			hasReady && readyCond.Status == core.ConditionFalse) {
			continue
		}

		if !leaders.Contains(state.Server(member.Member.ID)) {
			continue
		}

		r.log.
			Str("member", member.Member.ID).
			Info("Scheduling DisableMemberMaintenance: member has maintenance enabled, is not Ready and leads failover-capable shards")
		plan = append(plan, actions.NewAction(api.ActionTypeDisableMemberMaintenance, member.Group, member.Member, "Disable maintenance: member not Ready and leads failover-capable shards"))
	}
	return plan
}
