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

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	sharedReconcile "github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/rotation"
	"github.com/arangodb/kube-arangodb/pkg/util/compare"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createHighPlan considers the given specification & status and creates a plan to get the status in line with the specification.
// If a plan already exists, the given plan is returned with false.
// Otherwise, the new plan is returned with a boolean true.
func (r *Reconciler) createHighPlan(ctx context.Context, apiObject k8sutil.APIObject,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus,
	builderCtx PlanBuilderContext) (api.Plan, api.BackOff, bool) {
	if !currentPlan.IsEmpty() {
		// Plan already exists, complete that first
		return currentPlan, nil, false
	}

	q := recoverPlanAppender(r.log, newPlanAppender(NewWithPlanBuilder(ctx, apiObject, spec, status, builderCtx), status.BackOff, currentPlan).
		ApplyIfEmpty(r.updateMemberPodTemplateSpec).
		ApplyIfEmpty(r.updateMemberPhasePlan).
		ApplyIfEmpty(r.createCleanOutPlan).
		ApplyIfEmpty(r.syncMemberStatus).
		ApplyIfEmpty(r.createSyncPlan).
		ApplyIfEmpty(r.updateMemberUpdateConditionsPlan).
		ApplyIfEmpty(r.updateMemberRotationConditionsPlan).
		ApplyIfEmpty(r.createMemberAllowUpgradeConditionPlan).
		ApplyIfEmpty(r.createMemberRecreationConditionsPlan).
		ApplyIfEmpty(r.createMemberPodSchedulingFailurePlan).
		ApplyIfEmpty(r.createRotateServerStoragePVCPendingResizeConditionPlan).
		ApplyIfEmpty(r.createChangeMemberArchPlan).
		ApplyIfEmpty(r.createRotateServerStorageResizePlanRuntime).
		ApplyIfEmpty(r.createTopologyMemberUpdatePlan).
		ApplyIfEmptyWithBackOff(LicenseCheck, 30*time.Second, r.updateClusterLicense).
		ApplyIfEmpty(r.createTopologyMemberConditionPlan).
		ApplyIfEmpty(r.updateMemberConditionTypeMemberVolumeUnschedulableCondition).
		ApplyIfEmpty(r.createRebalancerCheckPlanCore).
		ApplyIfEmpty(r.createMemberFailedRestoreHighPlan).
		ApplyIfEmpty(r.volumeMemberReplacement).
		ApplyWithBackOff(BackOffCheck, time.Minute, r.emptyPlanBuilder)).
		ApplyIfEmptyWithBackOff(TimezoneCheck, time.Minute, r.createTimezoneUpdatePlan).
		Apply(r.createBackupInProgressConditionPlan). // Discover backups always
		Apply(r.createMaintenanceConditionPlan).      // Discover maintenance always
		Apply(r.cleanupConditions)                    // Cleanup Conditions

	return q.Plan(), q.BackOff(), true
}

// updateMemberPodTemplateSpec creates plan to update member Spec
func (r *Reconciler) updateMemberPodTemplateSpec(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	// Update member specs
	for _, e := range status.Members.AsList() {
		if e.Member.Phase != api.MemberPhaseNone {
			if reason, changed := r.arangoMemberPodTemplateNeedsUpdate(ctx, apiObject, spec, e.Group, status, e.Member, context); changed {
				plan = append(plan, actions.NewAction(api.ActionTypeArangoMemberUpdatePodSpec, e.Group, e.Member, reason))
			}
		}
	}

	return plan
}

// syncMemberStatus creates plan to sync member status
func (r *Reconciler) syncMemberStatus(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	for _, e := range status.Members.AsList() {
		cache, ok := context.ACS().ClusterCache(e.Member.ClusterID)
		if !ok {
			r.log.Error("Unable to get cache")
			continue
		}

		name := e.Member.ArangoMemberName(context.GetName(), e.Group)

		amember, ok := cache.ArangoMember().V1().GetSimple(name)
		if !ok {
			r.log.Error("Unable to get member")
			continue
		}

		if !amember.Status.InSync(e.Member) {
			plan = append(plan, actions.NewAction(api.ActionTypeMemberStatusSync, e.Group, e.Member, "Sync Status of ArangoMember"))
		}
	}

	return plan
}

// updateMemberPhasePlan creates plan to update member phase
func (r *Reconciler) updateMemberPhasePlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	for _, e := range status.Members.AsList() {
		if e.Member.Phase == api.MemberPhaseNone {
			var p api.Plan
			p = append(p,
				actions.NewAction(api.ActionTypeArangoMemberUpdatePodSpec, e.Group, e.Member, "Propagating spec of pod"),
				actions.NewAction(api.ActionTypeArangoMemberUpdatePodStatus, e.Group, e.Member, "Propagating status of pod"),
				actions.NewAction(api.ActionTypeMemberPhaseUpdate, e.Group, e.Member,
					"Move to Pending phase").AddParam(actionTypeMemberPhaseUpdatePhaseKey, api.MemberPhasePending.String()),
			)
			plan = append(plan, p...)
		}

		if e.Member.Phase == api.MemberPhaseCreationFailed {
			var p api.Plan
			p = append(p,
				actions.NewAction(api.ActionTypeMemberPhaseUpdate, e.Group, e.Member,
					"Move to None phase due to Creation Error").AddParam(actionTypeMemberPhaseUpdatePhaseKey, api.MemberPhaseNone.String()),
			)
			plan = append(plan, p...)
		}
	}

	return plan
}

func pendingRestartMemberConditionAction(group api.ServerGroup, memberID string, reason string) api.Action {
	return sharedReconcile.UpdateMemberConditionActionV2(reason, api.ConditionTypePendingRestart, group, memberID, true, reason, "", "")
}

func restartMemberConditionAction(group api.ServerGroup, memberID string, reason string) api.Plan {
	return api.Plan{
		pendingRestartMemberConditionAction(group, memberID, reason),
		sharedReconcile.UpdateMemberConditionActionV2(reason, api.ConditionTypeRestart, group, memberID, true, reason, "", ""),
	}
}

func tlsRotateConditionAction(group api.ServerGroup, memberID string, reason string) api.Action {
	return sharedReconcile.UpdateMemberConditionActionV2(reason, api.ConditionTypePendingTLSRotation, group, memberID, true, reason, "", "")
}

func (r *Reconciler) updateMemberUpdateConditionsPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	for _, e := range status.Members.AsList() {
		if e.Member.Conditions.IsTrue(api.ConditionTypeUpdating) {
			// We are in updating phase
			if status.Plan.IsEmpty() {
				// If plan is empty then something went wrong
				plan = append(plan, sharedReconcile.RemoveMemberConditionActionV2("Clean update actions after failure", api.ConditionTypePendingUpdate, e.Group, e.Member.ID),
					sharedReconcile.RemoveMemberConditionActionV2("Clean update actions after failure", api.ConditionTypeUpdating, e.Group, e.Member.ID),
					sharedReconcile.UpdateMemberConditionActionV2("Clean update actions after failure", api.ConditionTypeUpdateFailed, e.Group, e.Member.ID, true, "Clean update actions after failure", "", ""),
					sharedReconcile.UpdateMemberConditionActionV2("Clean update actions after failure", api.ConditionTypePendingRestart, e.Group, e.Member.ID, true, "Clean update actions after failure", "", ""))

			}
		}
	}

	return plan
}

func (r *Reconciler) updateMemberRotationConditionsPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	for _, e := range status.Members.AsList() {
		cache, ok := context.ACS().ClusterCache(e.Member.ClusterID)
		if !ok {
			continue
		}

		p, ok := cache.Pod().V1().GetSimple(e.Member.Pod.GetName())
		if !ok {
			p = nil
		}

		if p, err := r.updateMemberRotationConditions(apiObject, spec, e.Member, e.Group, p, context); err != nil {
			r.log.Err(err).Error("Error while generating rotation plan")
			return nil
		} else if len(p) > 0 {
			plan = append(plan, p...)
		}
	}

	return plan
}

func (r *Reconciler) updateMemberRotationConditions(apiObject k8sutil.APIObject, spec api.DeploymentSpec, member api.MemberStatus, group api.ServerGroup, p *core.Pod, context PlanBuilderContext) (api.Plan, error) {
	if member.Conditions.IsTrue(api.ConditionTypeRestart) {
		return nil, nil
	}

	arangoMember, ok := context.ACS().CurrentClusterCache().ArangoMember().V1().GetSimple(member.ArangoMemberName(apiObject.GetName(), group))
	if !ok {
		return nil, nil
	}

	if m, plan, checksum, reason, err := rotation.IsRotationRequired(context.ACS(), spec, member, group, p, arangoMember.Spec.Template, arangoMember.Status.Template); err != nil {
		r.log.Err(err).Error("Error while getting rotation details")
		return nil, err
	} else {
		switch m {
		case compare.EnforcedRotation:
			if reason != "" {
				r.log.Bool("enforced", true).Info(reason)
			} else {
				r.log.Bool("enforced", true).Info("Unknown reason")
			}
			// We need to do enforced rotation
			return restartMemberConditionAction(group, member.ID, reason), nil
		case compare.InPlaceRotation:
			if member.Conditions.IsTrue(api.ConditionTypeUpdateFailed) {
				if !(member.Conditions.IsTrue(api.ConditionTypePendingRestart) || member.Conditions.IsTrue(api.ConditionTypeRestart)) {
					return api.Plan{pendingRestartMemberConditionAction(group, member.ID, reason)}, nil
				}
				return nil, nil
			} else if member.Conditions.IsTrue(api.ConditionTypeUpdating) || member.Conditions.IsTrue(api.ConditionTypePendingUpdate) {
				return nil, nil
			}
			return api.Plan{sharedReconcile.UpdateMemberConditionActionV2(reason, api.ConditionTypePendingUpdate, group, member.ID, true, reason, "", "")}, nil
		case compare.SilentRotation:
			// Propagate changes without restart, but apply plan if required
			plan = append(plan, actions.NewAction(api.ActionTypeArangoMemberUpdatePodStatus, group, member, "Propagating status of pod").AddParam(ActionTypeArangoMemberUpdatePodStatusChecksum, checksum))
			return plan, nil
		case compare.GracefulRotation:
			if reason != "" {
				r.log.Bool("enforced", false).Info(reason)
			} else {
				r.log.Bool("enforced", false).Info("Unknown reason")
			}
			// We need to do graceful rotation
			if member.Conditions.IsTrue(api.ConditionTypePendingRestart) {
				return restartMemberConditionAction(group, member.ID, reason), nil
			}

			switch a := spec.MemberPropagationMode.Get(); a {
			case api.DeploymentMemberPropagationModeOnRestart:
				return api.Plan{pendingRestartMemberConditionAction(group, member.ID, reason)}, nil
			case api.DeploymentMemberPropagationModeAlways:
				return restartMemberConditionAction(group, member.ID, reason), nil
			default:
				r.log.Str("mode", a.String()).Warn("Unknown DeploymentMemberPropagationMode, fallback to Default")
				return restartMemberConditionAction(group, member.ID, reason), nil
			}
		default:
			return nil, nil
		}
	}
}
