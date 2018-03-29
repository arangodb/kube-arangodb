//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package resources

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	inspectedPodCounter = metrics.MustRegisterCounter("deployment", "inspected_pods", "Number of pod inspections")
)

const (
	podScheduleTimeout = time.Minute // How long we allow the schedule to take scheduling a pod.
)

// InspectPods lists all pods that belong to the given deployment and updates
// the member status of the deployment accordingly.
func (r *Resources) InspectPods() error {
	log := r.log
	var events []*v1.Event

	pods, err := r.context.GetOwnedPods()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get owned pods")
		return maskAny(err)
	}

	// Update member status from all pods found
	status := r.context.GetStatus()
	apiObject := r.context.GetAPIObject()
	var podNamesWithScheduleTimeout []string
	var unscheduledPodNames []string
	for _, p := range pods {
		if k8sutil.IsArangoDBImageIDAndVersionPod(p) {
			// Image ID pods are not relevant to inspect here
			continue
		}

		// Pod belongs to this deployment, update metric
		inspectedPodCounter.Inc()

		// Find member status
		memberStatus, group, found := status.Members.MemberStatusByPodName(p.GetName())
		if !found {
			log.Debug().Str("pod", p.GetName()).Msg("no memberstatus found for pod")
			continue
		}

		// Update state
		updateMemberStatusNeeded := false
		if k8sutil.IsPodSucceeded(&p) {
			// Pod has terminated with exit code 0.
			if memberStatus.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Succeeded", "") {
				log.Debug().Str("pod-name", p.GetName()).Msg("Updating member condition Terminated to true: Pod Succeeded")
				updateMemberStatusNeeded = true
				// Record termination time
				now := metav1.Now()
				memberStatus.RecentTerminations = append(memberStatus.RecentTerminations, now)
			}
		} else if k8sutil.IsPodFailed(&p) {
			// Pod has terminated with at least 1 container with a non-zero exit code.
			if memberStatus.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Failed", "") {
				log.Debug().Str("pod-name", p.GetName()).Msg("Updating member condition Terminated to true: Pod Failed")
				updateMemberStatusNeeded = true
				// Record termination time
				now := metav1.Now()
				memberStatus.RecentTerminations = append(memberStatus.RecentTerminations, now)
			}
		}
		if k8sutil.IsPodReady(&p) {
			// Pod is now ready
			if memberStatus.Conditions.Update(api.ConditionTypeReady, true, "Pod Ready", "") {
				log.Debug().Str("pod-name", p.GetName()).Msg("Updating member condition Ready to true")
				updateMemberStatusNeeded = true
			}
		} else {
			// Pod is not ready
			if memberStatus.Conditions.Update(api.ConditionTypeReady, false, "Pod Not Ready", "") {
				log.Debug().Str("pod-name", p.GetName()).Msg("Updating member condition Ready to false")
				updateMemberStatusNeeded = true
			}
		}
		if k8sutil.IsPodNotScheduledFor(&p, podScheduleTimeout) {
			// Pod cannot be scheduled for to long
			log.Debug().Str("pod-name", p.GetName()).Msg("Pod scheduling timeout")
			podNamesWithScheduleTimeout = append(podNamesWithScheduleTimeout, p.GetName())
		} else if !k8sutil.IsPodScheduled(&p) {
			unscheduledPodNames = append(unscheduledPodNames, p.GetName())
		}
		if updateMemberStatusNeeded {
			if err := status.Members.UpdateMemberStatus(memberStatus, group); err != nil {
				return maskAny(err)
			}
		}
	}

	podExists := func(podName string) bool {
		for _, p := range pods {
			if p.GetName() == podName {
				return true
			}
		}
		return false
	}

	// Go over all members, check for missing pods
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members *api.MemberStatusList) error {
		for _, m := range *members {
			if podName := m.PodName; podName != "" {
				if !podExists(podName) {
					switch m.State {
					case api.MemberStateNone:
						// Do nothing
					case api.MemberStateShuttingDown, api.MemberStateRotating, api.MemberStateUpgrading:
						// Shutdown was intended, so not need to do anything here.
						// Just mark terminated
						if m.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Terminated", "") {
							if err := status.Members.UpdateMemberStatus(m, group); err != nil {
								return maskAny(err)
							}
						}
					default:
						log.Debug().Str("pod-name", podName).Msg("Pod is gone")
						m.State = api.MemberStateNone // This is trigger a recreate of the pod.
						// Create event
						events = append(events, k8sutil.NewPodGoneEvent(podName, group.AsRole(), apiObject))
						if m.Conditions.Update(api.ConditionTypeReady, false, "Pod Does Not Exist", "") {
							if err := status.Members.UpdateMemberStatus(m, group); err != nil {
								return maskAny(err)
							}
						}
					}
				}
			}
		}
		return nil
	})

	// Update overall conditions
	allMembersReady := status.Members.AllMembersReady()
	status.Conditions.Update(api.ConditionTypeReady, allMembersReady, "", "")

	// Update conditions
	if len(podNamesWithScheduleTimeout) > 0 {
		if status.Conditions.Update(api.ConditionTypePodSchedulingFailure, true,
			"Pods Scheduling Timeout",
			fmt.Sprintf("The following pods cannot be scheduled: %v", podNamesWithScheduleTimeout)) {
			r.context.CreateEvent(k8sutil.NewPodsSchedulingFailureEvent(podNamesWithScheduleTimeout, r.context.GetAPIObject()))
		}
	} else if status.Conditions.IsTrue(api.ConditionTypePodSchedulingFailure) &&
		len(unscheduledPodNames) == 0 {
		if status.Conditions.Update(api.ConditionTypePodSchedulingFailure, false,
			"Pods Scheduling Resolved",
			"No pod reports a scheduling timeout") {
			r.context.CreateEvent(k8sutil.NewPodsSchedulingResolvedEvent(r.context.GetAPIObject()))
		}
	}

	// Save status
	if err := r.context.UpdateStatus(status); err != nil {
		return maskAny(err)
	}

	// Create events
	for _, evt := range events {
		r.context.CreateEvent(evt)
	}
	return nil
}
