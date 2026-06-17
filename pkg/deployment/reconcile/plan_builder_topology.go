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
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	sharedReconcile "github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/topology"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	actionTypeTopologyMemberAssignmentOperationRemove = "remove"
	actionTypeTopologyMemberAssignmentOperationAdd    = "add"
)

func (r *Reconciler) createTopologyEnablementPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	if spec.GetMode() == api.DeploymentModeSingle {
		// Topology cannot be changed in single server deployments
		return nil
	}

	if status.Topology == nil && spec.Topology.IsEnabled() {
		plan = append(plan, api.NewAction(api.ActionTypeTopologyEnable, api.ServerGroupUnknown, ""))
	} else if status.Topology != nil && !spec.Topology.IsEnabled() {
		plan = append(plan, api.NewAction(api.ActionTypeTopologyDisable, api.ServerGroupUnknown, ""))
	}

	return plan
}

func (r *Reconciler) createTopologyUpdatePlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	if spec.GetMode() == api.DeploymentModeSingle {
		// Topology cannot be changed in single server deployments
		return nil
	}

	if !status.Topology.Enabled() {
		return nil
	}

	mapping := getTopologyMappingObject(status)

	if !equality.Semantic.DeepEqual(status.Topology.Zones, mapping) {
		plan = append(plan, api.NewAction(api.ActionTypeTopologyZonesUpdate, api.ServerGroupUnknown, ""))
	}

	return plan
}

func (r *Reconciler) createTopologyMemberAdjustmentPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	if !status.Topology.Enabled() {
		return nil
	}

	// Check if all members are in place
	for id, zone := range status.Topology.Zones {
		for role, members := range zone.Members {
			gr := api.ServerGroupFromAbbreviatedRole(role)
			for _, member := range members {
				if _, g, ok := status.Members.ElementByID(member); !ok || g != gr {
					plan = append(plan, api.NewAction(api.ActionTypeTopologyMemberAssignment, gr, member).
						AddParam(actionTypeTopologyMemberAssignmentZone, fmt.Sprintf("%d", id)).
						AddParam(actionTypeTopologyMemberAssignmentID, string(status.Topology.ID)).
						AddParam(actionTypeTopologyMemberAssignmentOperation, actionTypeTopologyMemberAssignmentOperationRemove))
				}
			}
		}
	}

	return plan
}

func (r *Reconciler) createTopologyMemberUpdatePlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	for _, m := range status.Members.AsList() {
		if !status.Topology.Enabled() {
			if m.Member.Topology != nil {
				plan = append(plan, api.NewAction(api.ActionTypeTopologyMemberAssignment, m.Group, m.Member.ID).
					AddParam(actionTypeTopologyMemberAssignmentOperation, actionTypeTopologyMemberAssignmentOperationRemove))
			}
			continue
		}

		if t := status.Topology; t.Enabled() {
			if m.Member.Topology != nil {
				// Topology is not nil, but still not owned (race-condition)
				if !t.IsTopologyOwned(m.Member.Topology) {
					plan = append(plan, api.NewAction(api.ActionTypeTopologyMemberAssignment, m.Group, m.Member.ID).
						AddParam(actionTypeTopologyMemberAssignmentOperation, actionTypeTopologyMemberAssignmentOperationRemove))
				}
				continue
			}

			if m.Member.Pod.GetName() == "" {
				continue
			}

			cache, ok := context.ACS().ClusterCache(m.Member.ClusterID)
			if !ok {
				continue
			}

			nodes, err := cache.Node().V1()
			if err != nil {
				return nil
			}

			pod, ok := cache.Pod().V1().GetSimple(m.Member.Pod.GetName())
			if !ok {
				continue
			}

			if pod.Spec.NodeName == "" {
				continue
			}

			node, ok := nodes.GetSimple(pod.Spec.NodeName)
			if !ok {
				continue
			}

			v := node.Labels[t.Label]
			if v == "" {
				continue
			}

			for zone, zoneSpec := range t.Zones {
				if zoneSpec.Labels.Contains(v) {
					// We have found assignment for member
					plan = append(plan, api.NewAction(api.ActionTypeTopologyMemberAssignment, m.Group, m.Member.ID).
						AddParam(actionTypeTopologyMemberAssignmentZone, fmt.Sprintf("%d", zone)).
						AddParam(actionTypeTopologyMemberAssignmentID, string(t.ID)).
						AddParam(actionTypeTopologyMemberAssignmentLabel, v).
						AddParam(actionTypeTopologyMemberAssignmentOperation, actionTypeTopologyMemberAssignmentOperationAdd))
				}
			}
		}
	}

	return plan
}

func (r *Reconciler) createTopologyMemberConditionPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var plan api.Plan

	if t := status.Topology; t.Enabled() {
		members := status.Members.AsList()
		allMembersTA := true
		for _, e := range members {
			m := e.Member
			group := e.Group
			valid := t.IsTopologyOwned(m.Topology) && t.IsTopologyEvenlyDistributed(group)

			c, ok := m.Conditions.Get(api.ConditionTypeTopologyAware)
			if !valid {
				allMembersTA = false
				if !ok || c.IsTrue() {
					plan = append(plan, sharedReconcile.UpdateMemberConditionActionV2("Topology awareness disabled", api.ConditionTypeTopologyAware, group, m.ID, false, "Topology awareness disabled", "", ""))
				}
			} else {
				if !ok || !c.IsTrue() {
					plan = append(plan, sharedReconcile.UpdateMemberConditionActionV2("Topology awareness enabled", api.ConditionTypeTopologyAware, group, m.ID, true, "Topology awareness enabled", "", ""))
				}
			}
		}

		if allMembersTA {
			// Spec is TA
			if !status.Conditions.IsTrue(api.ConditionTypeTopologyAware) {
				plan = append(plan, sharedReconcile.UpdateConditionActionV2("Deployment is Topology Aware", api.ConditionTypeTopologyAware, true, "Deployment is Topology Aware", "", ""))
			}
		} else {
			if c, ok := status.Conditions.Get(api.ConditionTypeTopologyAware); !ok || c.IsTrue() {
				plan = append(plan, sharedReconcile.UpdateConditionActionV2("Deployment is not Topology Aware", api.ConditionTypeTopologyAware, false, "Deployment is not Topology Aware", "", ""))
			}
		}
	} else {
		for _, e := range status.Members.AsList() {
			if _, ok := e.Member.Conditions.Get(api.ConditionTypeTopologyAware); ok {
				plan = append(plan, sharedReconcile.RemoveMemberConditionActionV2("Cleaning Topology condition", api.ConditionTypeTopologyAware, e.Group, e.Member.ID))
			}
		}
		if _, ok := status.Conditions.Get(api.ConditionTypeTopologyAware); ok {
			plan = append(plan, sharedReconcile.RemoveConditionActionV2("Cleaning Topology condition", api.ConditionTypeTopologyAware))
		}
	}

	return plan
}

func getTopologyMappingObject(status api.DeploymentStatus) api.TopologyStatusZones {
	if !status.Topology.Enabled() {
		return nil
	}

	z := status.Topology.Zones.DeepCopy()

	for id := range z {
		z[id].Labels = nil
	}

	m, err := topology.GetTopologyMapping(status)
	if err != nil {
		logger.Warn("Multi assignment of the zones")
		return status.Topology.Zones.DeepCopy()
	}

	for k, v := range m {
		z[k].Labels = v
	}

	return z
}
