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
	"reflect"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createMemberPodSchedulingFailurePlan creates plan actions which are required when
// some pod has failed to schedule and scheduling parameters already changed
func (r *Reconciler) createMemberPodSchedulingFailurePlan(ctx context.Context,
	_ k8sutil.APIObject, spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) api.Plan {

	var p api.Plan
	if !status.Conditions.IsTrue(api.ConditionTypePodSchedulingFailure) {
		return p
	}

	for _, m := range status.Members.AsList() {
		l := r.log.Str("id", m.Member.ID).Str("role", m.Group.AsRole())

		if m.Member.Phase != api.MemberPhaseCreated || m.Member.Pod.GetName() == "" {
			// Act only when phase is created
			continue
		}

		if m.Member.Conditions.IsTrue(api.ConditionTypeScheduled) || m.Member.Conditions.IsTrue(api.ConditionTypeTerminating) {
			// Action is needed only for pods which are not scheduled yet
			continue
		}

		imageInfo, imageFound := context.SelectImageForMember(spec, status, m.Member)
		if !imageFound {
			l.Warn("could not find image for already created member")
			continue
		}

		renderedPod, err := context.RenderPodForMember(ctx, context.ACS(), spec, status, m.Member.ID, imageInfo)
		if err != nil {
			l.Err(err).Warn("could not render pod for already created member")
			continue
		}

		if r.isSchedulingParametersChanged(renderedPod.Spec, m.Member, context) {
			l.Info("Adding KillMemberPod action: scheduling failed and parameters already updated")
			p = append(p,
				actions.NewAction(api.ActionTypeKillMemberPod, m.Group, m.Member, "Scheduling failed"),
			)
		}
	}

	return p
}

// isSchedulingParametersChanged returns true if parameters related to pod scheduling has changed
func (r *Reconciler) isSchedulingParametersChanged(expectedSpec core.PodSpec, member api.MemberStatus, context PlanBuilderContext) bool {
	cache, ok := context.ACS().ClusterCache(member.ClusterID)
	if !ok {
		return false
	}
	pod, ok := cache.Pod().V1().GetSimple(member.Pod.GetName())
	if !ok {
		return false
	}
	if r.schedulingParametersAreTheSame(expectedSpec, pod.Spec) {
		return false
	}
	return true
}

func (r *Reconciler) schedulingParametersAreTheSame(expectedSpec, actualSpec core.PodSpec) bool {
	if expectedSpec.PriorityClassName != actualSpec.PriorityClassName {
		return false
	}

	if !reflect.DeepEqual(expectedSpec.Tolerations, actualSpec.Tolerations) {
		return false
	}

	if !reflect.DeepEqual(expectedSpec.NodeSelector, actualSpec.NodeSelector) {
		return false
	}

	// we should use SHA256 here because DeepEqual might be unreliable for Affinity rules
	if specC, err := util.SHA256FromJSON(expectedSpec.Affinity); err != nil {
		return true
	} else {
		if statusC, err := util.SHA256FromJSON(actualSpec.Affinity); err != nil {
			return true
		} else if specC != statusC {
			return false
		}
	}

	return true
}
