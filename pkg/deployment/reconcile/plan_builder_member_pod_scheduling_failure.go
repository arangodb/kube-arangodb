//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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
	"time"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createMemberPodSchedulingFailurePlan creates plan actions which are required when
// some pod has failed to schedule and scheduling parameters already changed
func (r *Reconciler) createMemberPodSchedulingFailurePlan(ctx context.Context,
	_ k8sutil.APIObject, spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) api.Plan {

	var p api.Plan

	if globals.GetGlobalTimeouts().PodSchedulingGracePeriod().Get() == 0 {
		// Scheduling grace period is not enabled
		return nil
	}

	if !status.Conditions.IsTrue(api.ConditionTypePodSchedulingFailure) {
		return p
	}

	q := r.log.Str("step", "CreateMemberPodSchedulingFailurePlan")

	for _, m := range status.Members.AsList() {
		l := q.Str("id", m.Member.ID).Str("role", m.Group.AsRole())

		if m.Member.Phase != api.MemberPhaseCreated || m.Member.Pod.GetName() == "" {
			// Act only when phase is created
			continue
		}

		if m.Member.Conditions.IsTrue(api.ConditionTypeScheduled) || m.Member.Conditions.IsTrue(api.ConditionTypeTerminating) {
			// Action is needed only for pods which are not scheduled yet
			continue
		}

		if c, ok := m.Member.Conditions.Get(api.ConditionTypeScheduled); !ok {
			// Action cant proceed if pod is not scheduled
			l.Debug("Unable to find scheduled condition")
			continue
		} else if c.LastTransitionTime.IsZero() {
			// LastTransitionTime is not set
			l.Debug("Scheduled condition LastTransitionTime is zero")
			continue
		} else {
			if d := time.Since(c.LastTransitionTime.Time); d <= globals.GetGlobalTimeouts().PodSchedulingGracePeriod().Get() {
				// In grace period
				l.Dur("since", d).Debug("Still in grace period")
				continue
			}
		}

		cache, ok := context.ACS().ClusterCache(m.Member.ClusterID)
		if !ok {
			l.Warn("Unable to get member name")
			continue
		}

		memberName := m.Member.ArangoMemberName(context.GetName(), m.Group)
		member, ok := cache.ArangoMember().V1().GetSimple(memberName)
		if !ok {
			l.Warn("Unable to get ArangoMember")
			continue
		}

		if m.Member.Conditions.IsTrue(api.ConditionTypeScheduleSpecChanged) {
			l.Info("Adding KillMemberPod action: scheduling failed and scheduling changed condition is present")
			p = append(p,
				actions.NewAction(api.ActionTypeKillMemberPod, m.Group, m.Member, "Scheduling failed"),
			)
		} else {
			if statusTemplate, specTemplate := member.Status.Template, member.Spec.Template; statusTemplate != nil && specTemplate != nil {
				if statusTemplateSpec, specTemplateSpec := statusTemplate.PodSpec, specTemplate.PodSpec; statusTemplateSpec != nil && specTemplateSpec != nil {
					if !r.schedulingParametersAreTheSame(specTemplateSpec.Spec, statusTemplateSpec.Spec) {
						l.Info("Adding KillMemberPod action: scheduling failed and parameters already updated")
						p = append(p,
							actions.NewAction(api.ActionTypeKillMemberPod, m.Group, m.Member, "Scheduling failed"),
						)
					} else {
						l.Info("Scheduling parameters are not updated")
					}
				} else {
					l.Warn("Pod TemplateSpec is nil")
				}
			} else {
				l.Warn("Pod Template is nil")
			}
		}
	}

	return p
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
