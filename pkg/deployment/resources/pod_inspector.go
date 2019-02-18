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
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	inspectedPodsCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_pods", "Number of pod inspections per deployment", metrics.DeploymentName)
	inspectPodsDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_pods_duration", "Amount of time taken by a single inspection of all pods for a deployment (in sec)", metrics.DeploymentName)
)

const (
	podScheduleTimeout              = time.Minute                // How long we allow the schedule to take scheduling a pod.
	recheckSoonPodInspectorInterval = util.Interval(time.Second) // Time between Pod inspection if we think something will change soon
	maxPodInspectorInterval         = util.Interval(time.Hour)   // Maximum time between Pod inspection (if nothing else happens)
)

// InspectPods lists all pods that belong to the given deployment and updates
// the member status of the deployment accordingly.
// Returns: Interval_till_next_inspection, error
func (r *Resources) InspectPods(ctx context.Context) (util.Interval, error) {
	log := r.log
	start := time.Now()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	var events []*k8sutil.Event
	nextInterval := maxPodInspectorInterval // Large by default, will be made smaller if needed in the rest of the function
	defer metrics.SetDuration(inspectPodsDurationGauges.WithLabelValues(deploymentName), start)

	pods, err := r.context.GetOwnedPods()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get owned pods")
		return 0, maskAny(err)
	}

	// Update member status from all pods found
	status, lastVersion := r.context.GetStatus()
	var podNamesWithScheduleTimeout []string
	var unscheduledPodNames []string
	for _, p := range pods {
		if k8sutil.IsArangoDBImageIDAndVersionPod(p) {
			// Image ID pods are not relevant to inspect here
			continue
		}

		// Pod belongs to this deployment, update metric
		inspectedPodsCounters.WithLabelValues(deploymentName).Inc()

		// Find member status
		memberStatus, group, found := status.Members.MemberStatusByPodName(p.GetName())
		if !found {
			log.Debug().Str("pod", p.GetName()).Msg("no memberstatus found for pod")
			if k8sutil.IsPodMarkedForDeletion(&p) && len(p.GetFinalizers()) > 0 {
				// Strange, pod belongs to us, but we have no member for it.
				// Remove all finalizers, so it can be removed.
				log.Warn().Msg("Pod belongs to this deployment, but we don't know the member. Removing all finalizers")
				kubecli := r.context.GetKubeCli()
				ignoreNotFound := false
				if err := k8sutil.RemovePodFinalizers(log, kubecli, &p, p.GetFinalizers(), ignoreNotFound); err != nil {
					log.Debug().Err(err).Msg("Failed to update pod (to remove all finalizers)")
					return 0, maskAny(err)
				}
			}
			continue
		}

		// Update state
		updateMemberStatusNeeded := false
		if k8sutil.IsPodSucceeded(&p) {
			// Pod has terminated with exit code 0.
			wasTerminated := memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated)
			if memberStatus.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Succeeded", "") {
				log.Debug().Str("pod-name", p.GetName()).Msg("Updating member condition Terminated to true: Pod Succeeded")
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
				if !wasTerminated {
					// Record termination time
					now := metav1.Now()
					memberStatus.RecentTerminations = append(memberStatus.RecentTerminations, now)
				}
			}
		} else if k8sutil.IsPodFailed(&p) {
			// Pod has terminated with at least 1 container with a non-zero exit code.
			wasTerminated := memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated)
			if memberStatus.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Failed", "") {
				log.Debug().Str("pod-name", p.GetName()).Msg("Updating member condition Terminated to true: Pod Failed")
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
				if !wasTerminated {
					// Record termination time
					now := metav1.Now()
					memberStatus.RecentTerminations = append(memberStatus.RecentTerminations, now)
				}
			}
		}
		if k8sutil.IsPodReady(&p) {
			// Pod is now ready
			if memberStatus.Conditions.Update(api.ConditionTypeReady, true, "Pod Ready", "") {
				log.Debug().Str("pod-name", p.GetName()).Msg("Updating member condition Ready to true")
				memberStatus.IsInitialized = true // Require future pods for this member to have an existing UUID (in case of dbserver).
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
			}
		} else {
			// Pod is not ready
			if memberStatus.Conditions.Update(api.ConditionTypeReady, false, "Pod Not Ready", "") {
				log.Debug().Str("pod-name", p.GetName()).Msg("Updating member condition Ready to false")
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
			}
		}
		if k8sutil.IsPodNotScheduledFor(&p, podScheduleTimeout) {
			// Pod cannot be scheduled for to long
			log.Debug().Str("pod-name", p.GetName()).Msg("Pod scheduling timeout")
			podNamesWithScheduleTimeout = append(podNamesWithScheduleTimeout, p.GetName())
		} else if !k8sutil.IsPodScheduled(&p) {
			unscheduledPodNames = append(unscheduledPodNames, p.GetName())
		}
		if k8sutil.IsPodMarkedForDeletion(&p) {
			// Process finalizers
			if x, err := r.runPodFinalizers(ctx, &p, memberStatus, func(m api.MemberStatus) error {
				updateMemberStatusNeeded = true
				memberStatus = m
				return nil
			}); err != nil {
				// Only log here, since we'll be called to try again.
				log.Warn().Err(err).Msg("Failed to run pod finalizers")
			} else {
				nextInterval = nextInterval.ReduceTo(x)
			}
		}
		if updateMemberStatusNeeded {
			if err := status.Members.Update(memberStatus, group); err != nil {
				return 0, maskAny(err)
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
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if podName := m.PodName; podName != "" {
				if !podExists(podName) {
					log.Debug().Str("pod-name", podName).Msg("Does not exist")
					switch m.Phase {
					case api.MemberPhaseNone:
						// Do nothing
						log.Debug().Str("pod-name", podName).Msg("PodPhase is None, waiting for the pod to be recreated")
					case api.MemberPhaseShuttingDown, api.MemberPhaseRotating, api.MemberPhaseUpgrading, api.MemberPhaseFailed:
						// Shutdown was intended, so not need to do anything here.
						// Just mark terminated
						wasTerminated := m.Conditions.IsTrue(api.ConditionTypeTerminated)
						if m.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Terminated", "") {
							if !wasTerminated {
								// Record termination time
								now := metav1.Now()
								m.RecentTerminations = append(m.RecentTerminations, now)
							}
							// Save it
							if err := status.Members.Update(m, group); err != nil {
								return maskAny(err)
							}
						}
					default:
						log.Debug().Str("pod-name", podName).Msg("Pod is gone")
						m.Phase = api.MemberPhaseNone // This is trigger a recreate of the pod.
						// Create event
						nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
						events = append(events, k8sutil.NewPodGoneEvent(podName, group.AsRole(), apiObject))
						updateMemberNeeded := false
						if m.Conditions.Update(api.ConditionTypeReady, false, "Pod Does Not Exist", "") {
							updateMemberNeeded = true
						}
						wasTerminated := m.Conditions.IsTrue(api.ConditionTypeTerminated)
						if m.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Does Not Exist", "") {
							if !wasTerminated {
								// Record termination time
								now := metav1.Now()
								m.RecentTerminations = append(m.RecentTerminations, now)
							}
							updateMemberNeeded = true
						}
						if updateMemberNeeded {
							// Save it
							if err := status.Members.Update(m, group); err != nil {
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
	spec := r.context.GetSpec()
	allMembersReady := status.Members.AllMembersReady(spec.GetMode(), spec.Sync.IsEnabled())
	status.Conditions.Update(api.ConditionTypeReady, allMembersReady, "", "")

	if spec.GetMode().HasCoordinators() && status.Members.Coordinators.AllFailed() {
		log.Error().Msg("All coordinators failed - reset")
		for _, m := range status.Members.Coordinators {
			if err := r.context.DeletePod(m.PodName); err != nil {
				log.Error().Err(err).Msg("Failed to delete pod")
			}
			m.Phase = api.MemberPhaseNone
			if err := status.Members.Update(m, api.ServerGroupCoordinators); err != nil {
				log.Error().Err(err).Msg("Failed to update member")
			}
		}
	}

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
	if err := r.context.UpdateStatus(status, lastVersion); err != nil {
		return 0, maskAny(err)
	}

	// Create events
	for _, evt := range events {
		r.context.CreateEvent(evt)
	}
	return nextInterval, nil
}

// GetExpectedPodArguments creates command line arguments for a server in the given group with given ID.
func (r *Resources) GetExpectedPodArguments(apiObject metav1.Object, deplSpec api.DeploymentSpec, group api.ServerGroup,
	agents api.MemberStatusList, id string, version driver.Version) []string {
	if group.IsArangod() {
		return createArangodArgs(apiObject, deplSpec, group, agents, id, version, false)
	}
	if group.IsArangosync() {
		groupSpec := deplSpec.GetServerGroupSpec(group)
		return createArangoSyncArgs(apiObject, deplSpec, group, groupSpec, agents, id)
	}
	return nil
}
