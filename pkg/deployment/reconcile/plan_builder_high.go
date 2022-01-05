//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package reconcile

import (
	"context"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/deployment/rotation"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
)

// createHighPlan considers the given specification & status and creates a plan to get the status in line with the specification.
// If a plan already exists, the given plan is returned with false.
// Otherwise the new plan is returned with a boolean true.
func createHighPlan(ctx context.Context, log zerolog.Logger, apiObject k8sutil.APIObject,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus, cachedStatus inspectorInterface.Inspector,
	builderCtx PlanBuilderContext) (api.Plan, api.BackOff, bool) {
	if !currentPlan.IsEmpty() {
		// Plan already exists, complete that first
		return currentPlan, nil, false
	}

	r := recoverPlanAppender(log, newPlanAppender(NewWithPlanBuilder(ctx, log, apiObject, spec, status, cachedStatus, builderCtx), status.BackOff, currentPlan).
		ApplyIfEmpty(updateMemberPodTemplateSpec).
		ApplyIfEmpty(updateMemberPhasePlan).
		ApplyIfEmpty(createCleanOutPlan).
		ApplyIfEmpty(updateMemberUpdateConditionsPlan).
		ApplyIfEmpty(updateMemberRotationConditionsPlan).
		ApplyIfEmpty(createTopologyMemberUpdatePlan).
		ApplyIfEmptyWithBackOff(LicenseCheck, 30*time.Second, updateClusterLicense).
		ApplyIfEmpty(createTopologyMemberConditionPlan).
		ApplyIfEmpty(createRebalancerCheckPlan).
		ApplyWithBackOff(BackOffCheck, time.Minute, emptyPlanBuilder))

	return r.Plan(), r.BackOff(), true
}

// updateMemberPodTemplateSpec creates plan to update member Spec
func updateMemberPodTemplateSpec(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	// Update member specs
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if m.Phase != api.MemberPhaseNone {
				if reason, changed := arangoMemberPodTemplateNeedsUpdate(ctx, log, apiObject, spec, group, status, m, cachedStatus, context); changed {
					plan = append(plan, api.NewAction(api.ActionTypeArangoMemberUpdatePodSpec, group, m.ID, reason))
				}
			}
		}

		return nil
	})

	return plan
}

// updateMemberPhasePlan creates plan to update member phase
func updateMemberPhasePlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			if m.Phase == api.MemberPhaseNone {
				var p api.Plan

				p = append(p,
					api.NewAction(api.ActionTypeArangoMemberUpdatePodSpec, group, m.ID, "Propagating spec of pod"),
					api.NewAction(api.ActionTypeArangoMemberUpdatePodStatus, group, m.ID, "Propagating status of pod"))

				p = append(p, api.NewAction(api.ActionTypeMemberPhaseUpdate, group, m.ID,
					"Move to Pending phase").AddParam(actionTypeMemberPhaseUpdatePhaseKey, api.MemberPhasePending.String()))

				plan = append(plan, p...)
			}
		}

		return nil
	})

	return plan
}

func pendingRestartMemberConditionAction(group api.ServerGroup, memberID string, reason string) api.Action {
	return api.NewAction(api.ActionTypeSetMemberCondition, group, memberID, reason).AddParam(api.ConditionTypePendingRestart.String(), "T")
}

func restartMemberConditionAction(group api.ServerGroup, memberID string, reason string) api.Action {
	return pendingRestartMemberConditionAction(group, memberID, reason).AddParam(api.ConditionTypeRestart.String(), "T")
}

func tlsRotateConditionAction(group api.ServerGroup, memberID string, reason string) api.Action {
	return api.NewAction(api.ActionTypeSetMemberCondition, group, memberID, reason).AddParam(api.ConditionTypePendingTLSRotation.String(), "T")
}

func updateMemberUpdateConditionsPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	if err := status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			if m.Conditions.IsTrue(api.ConditionTypeUpdating) {
				// We are in updating phase
				if status.Plan.IsEmpty() {
					// If plan is empty then something went wrong
					plan = append(plan,
						api.NewAction(api.ActionTypeSetMemberCondition, group, m.ID, "Clean update actions after failure").
							AddParam(api.ConditionTypePendingUpdate.String(), "").
							AddParam(api.ConditionTypeUpdating.String(), "").
							AddParam(api.ConditionTypeUpdateFailed.String(), "T").
							AddParam(api.ConditionTypePendingRestart.String(), "T"),
					)
				}
			}
		}

		return nil
	}); err != nil {
		log.Err(err).Msgf("Error while generating update plan")
		return nil
	}

	return plan
}

func updateMemberRotationConditionsPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	if err := status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			p, ok := cachedStatus.Pod(m.PodName)
			if !ok {
				p = nil
			}

			if p, err := updateMemberRotationConditions(log, apiObject, spec, cachedStatus, m, group, p); err != nil {
				return err
			} else if len(p) > 0 {
				plan = append(plan, p...)
			}
		}

		return nil
	}); err != nil {
		log.Err(err).Msgf("Error while generating rotation plan")
		return nil
	}

	return plan
}

func updateMemberRotationConditions(log zerolog.Logger, apiObject k8sutil.APIObject, spec api.DeploymentSpec, cachedStatus inspectorInterface.Inspector, member api.MemberStatus, group api.ServerGroup, p *core.Pod) (api.Plan, error) {
	if member.Conditions.IsTrue(api.ConditionTypeRestart) {
		return nil, nil
	}

	arangoMember, ok := cachedStatus.ArangoMember(member.ArangoMemberName(apiObject.GetName(), group))
	if !ok {
		return nil, nil
	}

	if m, _, reason, err := rotation.IsRotationRequired(log, cachedStatus, spec, member, group, p, arangoMember.Spec.Template, arangoMember.Status.Template); err != nil {
		log.Error().Err(err).Msgf("Error while getting rotation details")
		return nil, err
	} else {
		switch m {
		case rotation.EnforcedRotation:
			if reason != "" {
				log.Info().Bool("enforced", true).Msgf(reason)
			} else {
				log.Info().Bool("enforced", true).Msgf("Unknown reason")
			}
			// We need to do enforced rotation
			return api.Plan{restartMemberConditionAction(group, member.ID, reason)}, nil
		case rotation.InPlaceRotation:
			if member.Conditions.IsTrue(api.ConditionTypeUpdateFailed) {
				if !(member.Conditions.IsTrue(api.ConditionTypePendingRestart) || member.Conditions.IsTrue(api.ConditionTypeRestart)) {
					return api.Plan{pendingRestartMemberConditionAction(group, member.ID, reason)}, nil
				}
				return nil, nil
			} else if member.Conditions.IsTrue(api.ConditionTypeUpdating) || member.Conditions.IsTrue(api.ConditionTypePendingUpdate) {
				return nil, nil
			}
			return api.Plan{api.NewAction(api.ActionTypeSetMemberCondition, group, member.ID, reason).AddParam(api.ConditionTypePendingUpdate.String(), "T")}, nil
		case rotation.SilentRotation:
			// Propagate changes without restart
			return api.Plan{api.NewAction(api.ActionTypeArangoMemberUpdatePodStatus, group, member.ID, "Propagating status of pod").AddParam(ActionTypeArangoMemberUpdatePodStatusChecksum, arangoMember.Spec.Template.GetChecksum())}, nil
		case rotation.GracefulRotation:
			if reason != "" {
				log.Info().Bool("enforced", false).Msgf(reason)
			} else {
				log.Info().Bool("enforced", false).Msgf("Unknown reason")
			}
			// We need to do graceful rotation
			if member.Conditions.IsTrue(api.ConditionTypePendingRestart) {
				return nil, nil
			}

			if spec.MemberPropagationMode.Get() == api.DeploymentMemberPropagationModeAlways {
				return api.Plan{restartMemberConditionAction(group, member.ID, reason)}, nil
			} else {
				return api.Plan{pendingRestartMemberConditionAction(group, member.ID, reason)}, nil
			}
		default:
			return nil, nil
		}
	}
}
