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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createMemberFailedRestoreNormalPlan returns only actions which are not recreate member.
func (r *Reconciler) createMemberFailedRestoreNormalPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) api.Plan {
	condition := func(a api.Action) bool {
		return a.Type != api.ActionTypeRecreateMember
	}

	return r.createMemberFailedRestoreInternal(ctx, apiObject, spec, status, context).Filter(condition)
}

// createMemberFailedRestoreHighPlan returns only recreate member actions.
func (r *Reconciler) createMemberFailedRestoreHighPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) api.Plan {
	condition := func(a api.Action) bool {
		return a.Type == api.ActionTypeRecreateMember
	}

	return r.createMemberFailedRestoreInternal(ctx, apiObject, spec, status, context).Filter(condition)
}

func (r *Reconciler) createMemberFailedRestoreInternal(_ context.Context, _ k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	// Fetch agency plan.
	agencyState, agencyOK := context.GetAgencyCache()

	// Check for members in failed state.
	for _, group := range api.AllServerGroups {
		members := status.Members.MembersOfGroup(group)
		failed := 0
		for _, m := range members {
			if m.Phase == api.MemberPhaseFailed {
				failed++
			}
		}

		for _, m := range members {
			if m.Phase != api.MemberPhaseFailed || len(plan) > 0 {
				continue
			}

			memberLog := r.log.Str("id", m.ID).Str("role", group.AsRole())

			if group == api.ServerGroupDBServers && spec.GetMode() == api.DeploymentModeCluster {
				if !agencyOK {
					// If agency is down DBServers should not be touched.
					memberLog.Info("Agency state is not present")
					continue
				}

				if agencyState.Target.CleanedServers.Contains(state.Server(m.ID)) {
					memberLog.Info("Member is CleanedOut")
					continue
				}

				if agencyState.Plan.Collections.IsDBServerLeader(state.Server(m.ID)) {
					memberLog.Info("Recreating leader DBServer - it cannot be removed gracefully")
					plan = append(plan, actions.NewAction(api.ActionTypeRecreateMember, group, m))

					continue
				}

				if c := spec.DBServers.GetCount(); c <= len(members)-failed {
					// There are more or equal alive members than current count. A member should not be recreated.
					continue
				}

				if agencyState.Plan.Collections.IsDBServerPresent(state.Server(m.ID)) {
					// DBServer still exists in agency plan! Will not be removed, but needs to be recreated.
					memberLog.Info("Recreating DBServer - it cannot be removed gracefully")
					plan = append(plan, actions.NewAction(api.ActionTypeRecreateMember, group, m))

					continue
				}
				// From here on, DBServer can be recreated.
			}

			switch group {
			case api.ServerGroupAgents:
				// For agents just recreate member do not rotate ID, do not remove PVC or service.
				memberLog.Info("Restoring old member. For agency members recreation of PVC is not supported - to prevent DataLoss")
				plan = append(plan, actions.NewAction(api.ActionTypeRecreateMember, group, m))
			case api.ServerGroupSingle:
				// Do not remove data for single.
				memberLog.Info("Restoring old member. Rotation for single servers is not safe")
				plan = append(plan, actions.NewAction(api.ActionTypeRecreateMember, group, m))
			default:
				if spec.GetAllowMemberRecreation(group) {
					memberLog.Info("Creating member replacement plan because member has failed")
					plan = append(plan,
						actions.NewAction(api.ActionTypeRemoveMember, group, m),
						actions.NewAction(api.ActionTypeAddMember, group, shared.WithPredefinedMember("")),
						actions.NewAction(api.ActionTypeWaitForMemberUp, group, shared.WithPredefinedMember(api.MemberIDPreviousAction)),
					)
				} else {
					memberLog.Info("Restoring old member. Recreation is disabled for group")
					plan = append(plan, actions.NewAction(api.ActionTypeRecreateMember, group, m))
				}
			}
		}
	}

	if len(plan) == 0 && !agencyOK {
		r.log.Warn("unable to build further plan without access to agency")
		plan = append(plan, actions.NewClusterAction(api.ActionTypeIdle))
	}

	return plan
}
