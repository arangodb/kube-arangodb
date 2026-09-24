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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// hibernateShutdownGroupOrder is the order in which server groups are shut down when hibernating: the
// reverse of the startup order, so agents (holding the agency) are stopped last.
var hibernateShutdownGroupOrder = []api.ServerGroup{
	api.ServerGroupGateways,
	api.ServerGroupSyncWorkers,
	api.ServerGroupSyncMasters,
	api.ServerGroupCoordinators,
	api.ServerGroupDBServers,
	api.ServerGroupSingle,
	api.ServerGroupAgents,
}

// hibernateWakeGroupOrder is the order in which server groups are woken up when dehibernating: the
// normal startup order, so agents (holding the agency) come up first.
var hibernateWakeGroupOrder = []api.ServerGroup{
	api.ServerGroupAgents,
	api.ServerGroupSingle,
	api.ServerGroupDBServers,
	api.ServerGroupCoordinators,
	api.ServerGroupSyncMasters,
	api.ServerGroupSyncWorkers,
	api.ServerGroupGateways,
}

// createHibernatePlan drives deployment hibernation from spec.hibernate. When hibernation is requested
// it enables maintenance mode and shuts members down group by group into the Hibernated phase; when it
// is cleared it wakes the members back up group by group and disables maintenance mode. It emits a
// single group of actions per reconcile round so each group finishes before the next one is touched.
func (r *Reconciler) createHibernatePlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
	if spec.IsHibernate() {
		return r.createHibernateEnablePlan(spec, status, planCtx)
	}

	return r.createDehibernatePlan(spec, status, planCtx)
}

// createHibernateEnablePlan enables maintenance mode (cluster only) and then shuts members down group
// by group, in reverse startup order. Once the deployment is fully hibernated it returns an Idle action
// so that no other (JWT/TLS/rotate/scale) actions run while the deployment is hibernated.
func (r *Reconciler) createHibernateEnablePlan(spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
	// Enable maintenance mode first (cluster only), before any member is shut down.
	if spec.Mode.Get() == api.DeploymentModeCluster {
		agencyState, ok := planCtx.GetAgencyCache()
		if !ok {
			r.log.Error("Unable to get agency cache, delaying hibernation")
			return nil
		}

		if !agencyState.Supervision.Maintenance.Exists() {
			r.log.Info("Enabling maintenance mode for hibernation")
			return api.Plan{actions.NewClusterAction(api.ActionTypeEnableMaintenance, "Hibernation requested")}
		}
	}

	// Shut members down group by group, one group per reconcile round.
	for _, group := range hibernateShutdownGroupOrder {
		var plan api.Plan

		for _, m := range status.Members.MembersOfGroup(group) {
			if m.Phase.IsHibernated() {
				continue
			}

			plan = append(plan, actions.NewAction(api.ActionTypeHibernateMember, group, m, "Hibernation requested"))
		}

		if len(plan) > 0 {
			return plan
		}
	}

	// Deployment is fully hibernated: occupy the plan with an Idle action so no other actions run.
	return api.Plan{actions.NewClusterAction(api.ActionTypeIdle, "Deployment hibernated")}
}

// createDehibernatePlan wakes members up group by group in startup order and, once all are awake,
// disables maintenance mode.
func (r *Reconciler) createDehibernatePlan(spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
	// Wake members up group by group, one group per reconcile round.
	for _, group := range hibernateWakeGroupOrder {
		var plan api.Plan

		for _, m := range status.Members.MembersOfGroup(group) {
			if !m.Phase.IsHibernated() {
				continue
			}

			plan = append(plan, actions.NewAction(api.ActionTypeDehibernateMember, group, m, "Dehibernation requested"))
		}

		if len(plan) > 0 {
			return plan
		}
	}

	// No member is hibernated anymore; disable maintenance mode if it is still enabled (cluster only).
	if spec.Mode.Get() == api.DeploymentModeCluster {
		agencyState, ok := planCtx.GetAgencyCache()
		if !ok {
			return nil
		}

		if agencyState.Supervision.Maintenance.Exists() {
			r.log.Info("Disabling maintenance mode after dehibernation")
			return api.Plan{actions.NewClusterAction(api.ActionTypeDisableMaintenance, "Dehibernation completed")}
		}
	}

	return nil
}
