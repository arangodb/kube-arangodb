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

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createRotateServerStorageResizePlan creates plan to resize storage
func (r *Reconciler) createRotateServerStorageResizePlanRuntime(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	return r.createRotateServerStorageResizePlanInternal(spec, status, context, api.PVCResizeModeRuntime)
}

func (r *Reconciler) createRotateServerStorageResizePlanRotate(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	return r.createRotateServerStorageResizePlanInternal(spec, status, context, api.PVCResizeModeRotate)
}

func (r *Reconciler) createRotateServerStorageResizePlanInternal(spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext, mode api.PVCResizeMode) api.Plan {
	var plan api.Plan

	for _, member := range status.Members.AsList() {
		cache, ok := context.ACS().ClusterCache(member.Member.ClusterID)
		if !ok {
			// Do not work without cache
			continue
		}
		if member.Member.Phase != api.MemberPhaseCreated {
			// Only make changes when phase is created
			continue
		}
		if member.Member.PersistentVolumeClaim.GetName() == "" {
			// Plan is irrelevant without PVC
			continue
		}
		groupSpec := spec.GetServerGroupSpec(member.Group)

		if groupSpec.VolumeResizeMode.Get() != mode {
			continue
		}

		if !plan.IsEmpty() && groupSpec.VolumeResizeMode.Get() == api.PVCResizeModeRotate {
			// Only 1 change at a time
			continue
		}

		// Load PVC
		pvc, exists := cache.PersistentVolumeClaim().V1().GetSimple(member.Member.PersistentVolumeClaim.GetName())
		if !exists {
			r.planLogger.
				Str("role", member.Group.AsRole()).
				Str("id", member.Member.ID).
				Warn("Failed to get PVC")
			continue
		}

		var res core.ResourceList
		if groupSpec.HasVolumeClaimTemplate() {
			res = groupSpec.GetVolumeClaimTemplate().Spec.Resources.Requests
		} else {
			res = groupSpec.Resources.Requests
		}
		if requestedSize, ok := res[core.ResourceStorage]; ok {
			if volumeSize, ok := pvc.Spec.Resources.Requests[core.ResourceStorage]; ok {
				cmp := volumeSize.Cmp(requestedSize)
				if cmp < 0 {
					// Here we need to do proper calculation
					plan = append(plan, r.pvcResizePlan(member.Group, member.Member, mode)...)
				}
			}
		}
	}

	return plan
}

func (r *Reconciler) createRotateServerStoragePVCPendingResizeConditionPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan
	for _, i := range status.Members.AsList() {
		if i.Member.PersistentVolumeClaim.GetName() == "" {
			continue
		}

		pvc, exists := context.ACS().CurrentClusterCache().PersistentVolumeClaim().V1().GetSimple(i.Member.PersistentVolumeClaim.GetName())
		if !exists {
			continue
		}

		pvcResizePending := k8sutil.IsPersistentVolumeClaimFileSystemResizePending(pvc)
		pvcResizePendingCond := i.Member.Conditions.IsTrue(api.ConditionTypePVCResizePending)

		if pvcResizePending != pvcResizePendingCond {
			if pvcResizePending {
				plan = append(plan, shared.UpdateMemberConditionActionV2("PVC Resize pending", api.ConditionTypePVCResizePending, i.Group, i.Member.ID, true, "PVC Resize pending", "", ""))
			} else {
				plan = append(plan, shared.RemoveMemberConditionActionV2("PVC Resize is done", api.ConditionTypePVCResizePending, i.Group, i.Member.ID))
			}
		}
	}

	return plan
}

func (r *Reconciler) pvcResizePlan(group api.ServerGroup, member api.MemberStatus, mode api.PVCResizeMode) api.Plan {
	switch mode {
	case api.PVCResizeModeRuntime:
		return api.Plan{
			actions.NewAction(api.ActionTypePVCResize, group, member),
		}
	case api.PVCResizeModeRotate:
		return withWaitForMember(api.Plan{
			actions.NewAction(api.ActionTypeResignLeadership, group, member),
			actions.NewAction(api.ActionTypeKillMemberPod, group, member),
			actions.NewAction(api.ActionTypeRotateStartMember, group, member),
			actions.NewAction(api.ActionTypePVCResize, group, member),
			actions.NewAction(api.ActionTypePVCResized, group, member),
			actions.NewAction(api.ActionTypeRotateStopMember, group, member),
		}, group, member)
	default:
		r.planLogger.Str("server-group", group.AsRole()).Str("mode", mode.String()).
			Error("Requested mode is not supported")
		return nil
	}
}
