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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	sharedReconcile "github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (r *Reconciler) createScaleUPMemberPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	return r.createScaleMemberPlan(ctx, apiObject, spec, status, context).Filter(filterScaleUP)
}

func (r *Reconciler) createScaleMemberPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {

	var plan api.Plan

	switch spec.GetMode() {
	case api.DeploymentModeSingle:
		// Never scale down
		plan = append(plan, r.createScalePlan(status, status.Members.Single, api.ServerGroupSingle, 1, context).Filter(filterScaleUP)...)
	case api.DeploymentModeActiveFailover:
		// Only scale agents & singles
		if a := status.Agency; a != nil && a.Size != nil {
			plan = append(plan, r.createScalePlan(status, status.Members.Agents, api.ServerGroupAgents, int(*a.Size), context).Filter(filterScaleUP)...)
		}
		plan = append(plan, r.createScalePlan(status, status.Members.Single, api.ServerGroupSingle, spec.Single.GetCount(), context)...)
	case api.DeploymentModeCluster:
		// Scale agents, dbservers, coordinators
		if a := status.Agency; a != nil && a.Size != nil {
			plan = append(plan, r.createScalePlan(status, status.Members.Agents, api.ServerGroupAgents, int(*a.Size), context).Filter(filterScaleUP)...)
		}
		plan = append(plan, r.createScalePlan(status, status.Members.DBServers, api.ServerGroupDBServers, spec.DBServers.GetCount(), context)...)
		plan = append(plan, r.createScalePlan(status, status.Members.Coordinators, api.ServerGroupCoordinators, spec.Coordinators.GetCount(), context)...)
	}
	if spec.GetMode().SupportsSync() {
		// Scale syncmasters & syncworkers
		if context.IsSyncEnabled() {
			plan = append(plan, r.createScalePlan(status, status.Members.SyncMasters, api.ServerGroupSyncMasters, spec.SyncMasters.GetCount(), context)...)
			plan = append(plan, r.createScalePlan(status, status.Members.SyncWorkers, api.ServerGroupSyncWorkers, spec.SyncWorkers.GetCount(), context)...)
		} else {
			plan = append(plan, r.createScalePlan(status, status.Members.SyncMasters, api.ServerGroupSyncMasters, 0, context)...)
			plan = append(plan, r.createScalePlan(status, status.Members.SyncWorkers, api.ServerGroupSyncWorkers, 0, context)...)
		}
	}
	if features.IsGatewayEnabled(spec) {
		plan = append(plan, r.createScalePlan(status, status.Members.Gateways, api.ServerGroupGateways, spec.Gateways.GetCount(), context)...)
	} else {
		plan = append(plan, r.createScalePlan(status, status.Members.Gateways, api.ServerGroupGateways, 0, context)...)
	}

	return plan
}

// createScalePlan creates a scaling plan for a single server group
func (r *Reconciler) createScalePlan(status api.DeploymentStatus, members api.MemberStatusList, group api.ServerGroup, count int, context PlanBuilderContext) api.Plan {
	var plan api.Plan
	if len(members) < count {
		// Scale up
		toAdd := count - len(members)
		for i := 0; i < toAdd; i++ {
			plan = append(plan, actions.NewAction(api.ActionTypeAddMember, group, sharedReconcile.WithPredefinedMember("")))
		}
		r.planLogger.
			Int("count", count).
			Int("actual-count", len(members)).
			Int("delta", toAdd).
			Str("role", group.AsRole()).
			Debug("Creating scale-up plan")
	} else if len(members) > count {
		// Note, we scale down 1 member at a time

		if m, err := planBuilderScaleDownFilter(context, status, group, members); err != nil {
			r.planLogger.Err(err).Str("role", group.AsRole()).Warn("Failed to select member to remove")
		} else {
			ready, message := groupReadyForRestart(context, status, m, group)
			if !ready {
				r.planLogger.Str("member", m.ID).Str("role", group.AsRole()).Str("message", message).Warn("Unable to ScaleDown member")
				return nil
			}

			r.planLogger.
				Str("member-id", m.ID).
				Str("phase", string(m.Phase)).
				Debug("Found member to remove")
			plan = append(plan, cleanOutMember(group, m)...)
			r.planLogger.
				Int("count", count).
				Int("actual-count", len(members)).
				Str("role", group.AsRole()).
				Str("member-id", m.ID).
				Debug("Creating scale-down plan")
		}
	}
	return plan
}

func (r *Reconciler) createReplaceMemberPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {

	var plan api.Plan

	// Replace is only allowed for Coordinators, DBServers & Agents
	for _, e := range status.Members.AsListInGroups(api.ServerGroupAgents, api.ServerGroupDBServers, api.ServerGroupCoordinators) {
		if !plan.IsEmpty() {
			break
		}

		member := e.Member
		group := e.Group

		if member.Conditions.IsTrue(api.ConditionTypeMarkedToRemove) {
			ready, message := groupReadyForRestart(context, status, member, group)
			if !ready {
				r.planLogger.Str("member", member.ID).Str("role", group.AsRole()).Str("message", message).Warn("Unable to recreate member")
				continue
			}

			switch group {
			case api.ServerGroupDBServers:
				if len(status.Members.DBServers) <= spec.DBServers.GetCount() {
					plan = append(plan,
						actions.NewAction(api.ActionTypeAddMember, group, sharedReconcile.WithPredefinedMember("")),
						actions.NewAction(api.ActionTypeDelay, group, sharedReconcile.WithPredefinedMember(api.MemberIDPreviousAction)).AddParam(DelayActionDuration, (15*time.Second).String()),
						actions.NewAction(api.ActionTypeMigrateMember, group, sharedReconcile.WithPredefinedMember(api.MemberIDPreviousAction)).AddParam(actionMigrateMemberSourceKey, member.ID),
					)
					r.planLogger.
						Str("role", group.AsRole()).
						Debug("Creating replacement plan")
				}
			case api.ServerGroupCoordinators:
				plan = append(plan, cleanOutMember(group, member)...)
				r.planLogger.
					Str("role", group.AsRole()).
					Debug("Creating replacement plan")
			case api.ServerGroupAgents:
				plan = append(plan, cleanOutMember(group, member)...)
				plan = append(plan, actions.NewAction(api.ActionTypeAddMember, group, sharedReconcile.WithPredefinedMember("")))
				r.planLogger.
					Str("role", group.AsRole()).
					Debug("Creating replacement plan")
			}
		}
	}

	return plan
}

func filterScaleUP(a api.Action) bool {
	return a.Type == api.ActionTypeAddMember
}
