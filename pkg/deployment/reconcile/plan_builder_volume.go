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

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	sharedApis "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

func (r *Reconciler) volumeMemberReplacement(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if !features.LocalVolumeReplacementCheck().Enabled() {
		return nil
	}

	cache, ok := context.GetAgencyCache()
	if !ok {
		// Cache is not ready
		return nil
	}

	servers := cache.PlanLeaderServers()

	for _, member := range status.Members.AsList() {
		if member.Member.Conditions.IsTrue(api.ConditionTypeScheduled) {
			continue
		}

		if !member.Member.Conditions.IsTrue(api.ConditionTypeMemberVolumeUnschedulable) {
			continue
		}

		if servers.Contains(agency.Server(member.Member.ID)) {
			continue
		}

		if pvc := member.Member.PersistentVolumeClaim; pvc != nil {
			if n := pvc.GetName(); n != "" {
				client, ok := context.ACS().ClusterCache(member.Member.ClusterID)
				if ok {
					if pvc, ok := client.PersistentVolumeClaim().V1().GetSimple(n); ok {
						// Server is not part of plan and is not ready
						return api.Plan{actions.NewAction(api.ActionTypeRemoveMemberPVC, member.Group, member.Member, "PVC is unschedulable").AddParam("pvc", string(pvc.GetUID()))}
					}
				}
			}
		}
	}

	return nil
}

// updateMemberPhasePlan creates plan to update member phase
func (r *Reconciler) updateMemberConditionTypeMemberVolumeUnschedulableCondition(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	cache := context.ACS().CurrentClusterCache()

	volumeClient, err := cache.PersistentVolume().V1()
	if err != nil {
		// We cant fetch volumes, continue
		return nil
	}

	for _, e := range status.Members.AsList() {
		if pvcStatus := e.Member.PersistentVolumeClaim; pvcStatus != nil {
			if pvc, ok := context.ACS().CurrentClusterCache().PersistentVolumeClaim().V1().GetSimple(pvcStatus.GetName()); ok {
				if volumeName := pvc.Spec.VolumeName; volumeName != "" {
					if pv, ok := volumeClient.GetSimple(volumeName); ok {
						// We have volume and volumeclaim, lets calculate condition
						unschedulable := memberConditionTypeMemberVolumeUnschedulableCalculate(cache, pv, pvc,
							memberConditionTypeMemberVolumeUnschedulableLocalStorageGone)

						if e.Member.Conditions.IsTrue(api.ConditionTypeScheduled) {
							// We are scheduled, above checks can be ignored
							unschedulable = false
						}

						if unschedulable == e.Member.Conditions.IsTrue(api.ConditionTypeMemberVolumeUnschedulable) {
							continue
						} else if unschedulable && !e.Member.Conditions.IsTrue(api.ConditionTypeMemberVolumeUnschedulable) {
							plan = append(plan, shared.UpdateMemberConditionActionV2("PV Unschedulable", api.ConditionTypeMemberVolumeUnschedulable, e.Group, e.Member.ID, true,
								"PV Unschedulable", "PV Unschedulable", ""))
						} else if !unschedulable && e.Member.Conditions.IsTrue(api.ConditionTypeMemberVolumeUnschedulable) {
							plan = append(plan, shared.RemoveMemberConditionActionV2("PV Schedulable", api.ConditionTypeMemberVolumeUnschedulable, e.Group, e.Member.ID))
						}
					}
				}
			}
		}
	}

	return plan
}

type memberConditionTypeMemberVolumeUnschedulableCalculateFunc func(cache inspectorInterface.Inspector, pv *core.PersistentVolume, pvc *core.PersistentVolumeClaim) bool

func memberConditionTypeMemberVolumeUnschedulableCalculate(cache inspectorInterface.Inspector, pv *core.PersistentVolume, pvc *core.PersistentVolumeClaim, funcs ...memberConditionTypeMemberVolumeUnschedulableCalculateFunc) bool {
	for _, f := range funcs {
		if f(cache, pv, pvc) {
			return true
		}
	}

	return false
}

func memberConditionTypeMemberVolumeUnschedulableLocalStorageGone(cache inspectorInterface.Inspector, pv *core.PersistentVolume, _ *core.PersistentVolumeClaim) bool {
	nodes, err := cache.Node().V1()
	if err != nil {
		return false
	}

	if pv.Spec.PersistentVolumeSource.Local == nil {
		// We are not on LocalStorage
		return false
	}

	if nodeAffinity := pv.Spec.NodeAffinity; nodeAffinity != nil {
		if required := nodeAffinity.Required; required != nil {
			for _, nst := range required.NodeSelectorTerms {
				for _, expr := range nst.MatchExpressions {
					if expr.Key == sharedApis.TopologyKeyHostname && expr.Operator == core.NodeSelectorOpIn {
						// We got exact key which is required for PV
						if len(expr.Values) == 1 {
							// Only one host assigned, we use it as localStorage - check if node exists
							_, ok := nodes.GetSimple(expr.Values[0])
							if !ok {
								// Node is missing!
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}
