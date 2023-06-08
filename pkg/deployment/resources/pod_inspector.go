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

package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/info"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

var (
	inspectedPodsCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_pods", "Number of pod inspections per deployment", metrics.DeploymentName)
	inspectPodsDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_pods_duration", "Amount of time taken by a single inspection of all pods for a deployment (in sec)", metrics.DeploymentName)
)

const (
	podScheduleTimeout       = time.Minute       // How long we allow the schedule to take scheduling a pod.
	terminationRestartPeriod = time.Second * -30 // If previous pod termination happened less than this time ago,
	// we will mark the pod as scheduled for termination
	recheckSoonPodInspectorInterval = util.Interval(time.Second) // Time between Pod inspection if we think something will change soon
	maxPodInspectorInterval         = util.Interval(time.Hour)   // Maximum time between Pod inspection (if nothing else happens)
	forcePodDeletionGracePeriod     = 15 * time.Minute
)

func (r *Resources) handleRestartedPod(pod *core.Pod, memberStatus *api.MemberStatus, wasTerminated, markAsTerminated *bool) {
	containerStatus, exist := k8sutil.GetContainerStatusByName(pod, api.ServerGroupReservedContainerNameServer)
	if exist && containerStatus.State.Terminated != nil {
		// do not record termination time again in the code below
		*wasTerminated = true

		termination := containerStatus.State.Terminated.FinishedAt
		if memberStatus.RecentTerminationsSince(termination.Time) == 0 {
			memberStatus.RecentTerminations = append(memberStatus.RecentTerminations, termination)
		}

		previousTermination := containerStatus.LastTerminationState.Terminated
		allowedRestartPeriod := time.Now().Add(terminationRestartPeriod)
		if previousTermination != nil && !previousTermination.FinishedAt.Time.Before(allowedRestartPeriod) {
			r.log.Str("pod-name", pod.GetName()).Debug("pod is continuously restarting - we will terminate it")
			*markAsTerminated = true
		} else {
			*markAsTerminated = false
			r.log.Str("pod-name", pod.GetName()).Debug("pod is restarting - we are not marking it as terminated yet..")
		}
	}
}

// InspectPods lists all pods that belong to the given deployment and updates
// the member status of the deployment accordingly.
// Returns: Interval_till_next_inspection, error
func (r *Resources) InspectPods(ctx context.Context, cachedStatus inspectorInterface.Inspector) (util.Interval, error) {
	log := r.log.Str("section", "pod")
	start := time.Now()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	var events []*k8sutil.Event
	nextInterval := maxPodInspectorInterval // Large by default, will be made smaller if needed in the rest of the function
	defer metrics.SetDuration(inspectPodsDurationGauges.WithLabelValues(deploymentName), start)

	agencyCache, agencyCachePresent := r.context.GetAgencyCache()

	status := r.context.GetStatus()
	var podNamesWithScheduleTimeout []string
	var unscheduledPodNames []string

	err := cachedStatus.Pod().V1().Iterate(func(pod *core.Pod) error {
		if info.GetPodServerGroup(pod) == api.ServerGroupImageDiscovery {
			// Image ID pods are not relevant to inspect here
			return nil
		}

		// Pod belongs to this deployment, update metric
		inspectedPodsCounters.WithLabelValues(deploymentName).Inc()

		memberStatus, group, found := status.Members.MemberStatusByPodName(pod.GetName())
		if !found {
			log.Str("pod", pod.GetName()).Strs("existing-pods", status.Members.PodNames()...).Warn("no memberstatus found for pod")
			if k8sutil.IsPodMarkedForDeletion(pod) && len(pod.GetFinalizers()) > 0 {
				// Strange, pod belongs to us, but we have no member for it.
				// Remove all finalizers, so it can be removed.
				log.Warn("Pod belongs to this deployment, but we don't know the member. Removing all finalizers")
				_, err := k8sutil.RemovePodFinalizers(ctx, r.context.ACS().CurrentClusterCache(), cachedStatus.PodsModInterface().V1(), pod, pod.GetFinalizers(), false)
				if err != nil {
					log.Err(err).Debug("Failed to update pod (to remove all finalizers)")
					return errors.WithStack(err)
				}
			}
			return nil
		}

		spec := r.context.GetSpec()
		groupSpec := spec.GetServerGroupSpec(group)
		coreContainers := spec.GetCoreContainers(group)

		if c, ok := memberStatus.Conditions.Get(api.ConditionTypeUpdating); ok {
			if v, ok := c.Params[api.ConditionParamContainerUpdatingName]; ok {
				// We are in update phase, container needs to be ignored
				if v != "" {
					coreContainers = coreContainers.Remove(v)
				}
			}
		} else {
			// Restore gracefulness
			if !k8sutil.IsPodTerminating(pod) {
				if err := k8sutil.EnsureFinalizerPresent(ctx, cachedStatus.PodsModInterface().V1(), pod, k8sutil.GetFinalizers(groupSpec, group)...); err != nil {
					log.Err(err).Error("Unable to enforce finalizer")
				}
			}
		}

		// Update state
		updateMemberStatusNeeded := false
		if k8sutil.IsPodSucceeded(pod, coreContainers) {
			// Pod has terminated with exit code 0.
			wasTerminated := memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated)
			markAsTerminated := true

			if pod.Spec.RestartPolicy == core.RestartPolicyAlways && !wasTerminated {
				r.handleRestartedPod(pod, &memberStatus, &wasTerminated, &markAsTerminated)
			}

			if markAsTerminated && memberStatus.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Succeeded", "") {
				log.Str("pod-name", pod.GetName()).Debug("Updating member condition Terminated to true: Pod Succeeded")
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)

				if !wasTerminated {
					// Record termination time
					now := meta.Now()
					memberStatus.RecentTerminations = append(memberStatus.RecentTerminations, now)
				}
			}
		} else if k8sutil.IsPodFailed(pod, coreContainers) {
			// Pod has terminated with at least 1 container with a non-zero exit code.
			wasTerminated := memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated)
			markAsTerminated := true

			if pod.Spec.RestartPolicy == core.RestartPolicyAlways && !wasTerminated {
				r.handleRestartedPod(pod, &memberStatus, &wasTerminated, &markAsTerminated)
			}

			if markAsTerminated && memberStatus.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Failed", "") {
				if containers := k8sutil.GetFailedContainerNames(pod.Status.InitContainerStatuses); len(containers) > 0 {
					for _, container := range containers {
						switch container {
						case api.ServerGroupReservedInitContainerNameVersionCheck:
							if c, ok := k8sutil.GetAnyContainerStatusByName(pod.Status.InitContainerStatuses, container); ok {
								if t := c.State.Terminated; t != nil && t.ExitCode == 11 {
									memberStatus.Upgrade = true
									updateMemberStatusNeeded = true
								}
							}
						case api.ServerGroupReservedInitContainerNameUpgrade:
							memberStatus.Conditions.Update(api.ConditionTypeUpgradeFailed, true, "Upgrade Failed", "")
						}

						if c, ok := k8sutil.GetAnyContainerStatusByName(pod.Status.InitContainerStatuses, container); ok {
							if t := c.State.Terminated; t != nil && t.ExitCode != 0 {
								log.Str("member", memberStatus.ID).
									Str("pod", pod.GetName()).
									Str("container", container).
									Str("uid", string(pod.GetUID())).
									Int32("exit-code", t.ExitCode).
									Str("reason", t.Reason).
									Str("message", t.Message).
									Int32("signal", t.Signal).
									Time("started", t.StartedAt.Time).
									Time("finished", t.FinishedAt.Time).
									Warn("Pod failed in unexpected way: Init Container failed")

								r.metrics.IncMemberInitContainerRestarts(memberStatus.ID, container, t.Reason, t.ExitCode)
							}
						}
					}
				}

				if containers := k8sutil.GetFailedContainerNames(pod.Status.ContainerStatuses); len(containers) > 0 {
					for _, container := range containers {
						if c, ok := k8sutil.GetAnyContainerStatusByName(pod.Status.ContainerStatuses, container); ok {
							if t := c.State.Terminated; t != nil && t.ExitCode != 0 {
								log.Str("member", memberStatus.ID).
									Str("pod", pod.GetName()).
									Str("container", container).
									Str("uid", string(pod.GetUID())).
									Int32("exit-code", t.ExitCode).
									Str("reason", t.Reason).
									Str("message", t.Message).
									Int32("signal", t.Signal).
									Time("started", t.StartedAt.Time).
									Time("finished", t.FinishedAt.Time).
									Warn("Pod failed in unexpected way: Core Container failed")

								r.metrics.IncMemberContainerRestarts(memberStatus.ID, container, t.Reason, t.ExitCode)
							}
						}
					}
				}

				log.Str("pod-name", pod.GetName()).Debug("Updating member condition Terminated to true: Pod Failed")
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)

				if !wasTerminated {
					// Record termination time
					now := meta.Now()
					memberStatus.RecentTerminations = append(memberStatus.RecentTerminations, now)
				}
			}
		}

		if k8sutil.IsPodScheduled(pod) {
			if _, ok := pod.Labels[k8sutil.LabelKeyArangoScheduled]; !ok {
				// Adding scheduled label to the pod
				l := addLabel(pod.Labels, k8sutil.LabelKeyArangoScheduled, "1")

				if err := r.context.ApplyPatchOnPod(ctx, pod, patch.ItemReplace(patch.NewPath("metadata", "labels"), l)); err != nil {
					log.Err(err).Error("Unable to update scheduled labels")
				}
			}
		}

		// Topology labels
		tv, tok := pod.Labels[k8sutil.LabelKeyArangoTopology]
		zv, zok := pod.Labels[k8sutil.LabelKeyArangoZone]

		if t, ts := status.Topology, memberStatus.Topology; t.Enabled() && t.IsTopologyOwned(ts) {
			if tid, tz := string(t.ID), fmt.Sprintf("%d", ts.Zone); !tok || !zok || tv != tid || zv != tz {
				l := addLabel(pod.Labels, k8sutil.LabelKeyArangoTopology, tid)
				l = addLabel(l, k8sutil.LabelKeyArangoZone, tz)

				if err := r.context.ApplyPatchOnPod(ctx, pod, patch.ItemReplace(patch.NewPath("metadata", "labels"), l)); err != nil {
					log.Err(err).Error("Unable to update topology labels")
				}
			}
		} else {
			if tok || zok {
				l := removeLabel(pod.Labels, k8sutil.LabelKeyArangoTopology)
				l = removeLabel(l, k8sutil.LabelKeyArangoZone)

				if err := r.context.ApplyPatchOnPod(ctx, pod, patch.ItemReplace(patch.NewPath("metadata", "labels"), l)); err != nil {
					log.Err(err).Error("Unable to remove topology labels")
				}
			}
		}
		// End of Topology labels

		// Reachable state
		if state, ok := r.context.GetMembersState().MemberState(memberStatus.ID); ok {
			if state.IsReachable() {
				if memberStatus.Conditions.Update(api.ConditionTypeReachable, true, "ArangoDB is reachable", "") {
					updateMemberStatusNeeded = true
					nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
				}
			} else {
				if memberStatus.Conditions.Update(api.ConditionTypeReachable, false, "ArangoDB is not reachable", "") {
					updateMemberStatusNeeded = true
					nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
				}
			}
		}

		// Member arch check
		if v, ok := pod.Annotations[deployment.ArangoDeploymentPodChangeArchAnnotation]; ok {
			if api.ArangoDeploymentArchitectureType(v).IsArchMismatch(spec.Architecture, memberStatus.Architecture) {
				if memberStatus.Conditions.Update(api.ConditionTypeArchitectureMismatch, true, "Member has a different architecture than the deployment", "") {
					updateMemberStatusNeeded = true
				}
			}
		}

		// Member Maintenance
		if agencyCachePresent {
			if agencyCache.Current.MaintenanceServers.InMaintenance(state.Server(memberStatus.ID)) {
				if memberStatus.Conditions.Update(api.ConditionTypeMemberMaintenanceMode, true, "ArangoDB Member maintenance enabled", "") {
					updateMemberStatusNeeded = true
					nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
				}
			} else {
				if memberStatus.Conditions.Remove(api.ConditionTypeMemberMaintenanceMode) {
					updateMemberStatusNeeded = true
					nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
				}
			}
		}

		if k8sutil.IsContainerStarted(pod, shared.ServerContainerName) {
			if memberStatus.Conditions.Update(api.ConditionTypeActive, true, "Core Pod Container started", "") {
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
			}
		}

		if memberStatus.Conditions.IsTrue(api.ConditionTypeActive) {
			if v, ok := pod.Labels[k8sutil.LabelKeyArangoActive]; !ok || v != k8sutil.LabelValueArangoActive {
				pod.Labels[k8sutil.LabelKeyArangoActive] = k8sutil.LabelValueArangoActive
				if err := r.context.ApplyPatchOnPod(ctx, pod, patch.ItemReplace(patch.NewPath("metadata", "labels"), pod.Labels)); err != nil {
					log.Str("pod-name", pod.GetName()).Err(err).Error("Unable to update labels")
				}
			}

		}

		if k8sutil.IsPodReady(pod) && k8sutil.AreContainersReady(pod, coreContainers) {
			// Pod is now ready
			if anyOf(memberStatus.Conditions.Update(api.ConditionTypeReady, true, "Pod Ready", ""),
				memberStatus.Conditions.Update(api.ConditionTypeStarted, true, "Pod Started", ""),
				memberStatus.Conditions.Update(api.ConditionTypeServing, true, "Pod Serving", "")) {
				log.Str("pod-name", pod.GetName()).Debug("Updating member condition Ready, Started & Serving to true")

				if status.Topology.IsTopologyOwned(memberStatus.Topology) {
					nodes, err := cachedStatus.Node().V1()
					if err == nil {
						node, ok := nodes.GetSimple(pod.Spec.NodeName)
						if ok {
							label, ok := node.Labels[status.Topology.Label]
							if ok {
								memberStatus.Topology.Label = label
							}
						}
					}
				}

				memberStatus.IsInitialized = true // Require future pods for this member to have an existing UUID (in case of dbserver).
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
			}
		} else if k8sutil.AreContainersReady(pod, coreContainers) {
			// Pod is not ready, but core containers are fine
			if anyOf(memberStatus.Conditions.Update(api.ConditionTypeReady, false, "Pod Not Ready", ""),
				memberStatus.Conditions.Update(api.ConditionTypeServing, true, "Pod is still serving", "")) {
				log.Str("pod-name", pod.GetName()).Debug("Updating member condition Ready to false, while all core containers are ready")
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
			}
		} else {
			// Pod is not ready
			if anyOf(memberStatus.Conditions.Update(api.ConditionTypeReady, false, "Pod Not Ready", ""),
				memberStatus.Conditions.Update(api.ConditionTypeServing, false, "Pod Core containers are not ready", strings.Join(coreContainers, ", "))) {
				log.Str("pod-name", pod.GetName()).Debug("Updating member condition Ready & Serving to false")
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
			}
		}

		if k8sutil.IsPodScheduled(pod) {
			if memberStatus.Conditions.Update(api.ConditionTypeScheduled, true, "Pod is scheduled", "") {
				updateMemberStatusNeeded = true
				nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
			}
		} else {
			if k8sutil.IsPodNotScheduledFor(pod, podScheduleTimeout) {
				// Pod cannot be scheduled for to long
				log.Str("pod-name", pod.GetName()).Debug("Pod scheduling timeout")
				podNamesWithScheduleTimeout = append(podNamesWithScheduleTimeout, pod.GetName())
			} else {
				unscheduledPodNames = append(unscheduledPodNames, pod.GetName())
			}
		}

		if k8sutil.IsPodMarkedForDeletion(pod) {
			if memberStatus.Conditions.Update(api.ConditionTypeTerminating, true, "Pod marked for deletion", "") {
				updateMemberStatusNeeded = true
				log.Str("pod-name", pod.GetName()).Debug("Pod marked as terminating")
			}

			// Process finalizers
			if x, err := r.runPodFinalizers(ctx, pod, memberStatus, func(m api.MemberStatus) error {
				updateMemberStatusNeeded = true
				memberStatus = m
				return nil
			}); err != nil {
				// Only log here, since we'll be called to try again.
				log.Err(err).Warn("Failed to run pod finalizers")
			} else {
				nextInterval = nextInterval.ReduceTo(x)
			}

			// Check if any additional deletion request is required
			if !k8sutil.IsPodAlive(pod) {
				var gps int64 = 10

				forceDelete := false
				if t := k8sutil.PodStopTime(pod); !t.IsZero() {
					if time.Since(t) > forcePodDeletionGracePeriod {
						forceDelete = true
					}
				} else if t := pod.DeletionTimestamp; t != nil {
					if time.Since(t.Time) > forcePodDeletionGracePeriod {
						forceDelete = true
					}
				}

				if forceDelete {
					gps = 0
					log.Str("pod-name", pod.GetName()).Warn("Enforcing deletion of Pod")
				}

				// Pod is dead, but still not removed. Send additional deletion request
				nctx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
				defer c()

				if err := cachedStatus.PodsModInterface().V1().Delete(nctx, pod.GetName(), meta.DeleteOptions{
					GracePeriodSeconds: util.NewType[int64](gps),
					Preconditions:      meta.NewUIDPreconditions(string(pod.GetUID())),
				}); err != nil {
					if kerrors.IsNotFound(err) {
						// Pod is already gone, we are fine with it
					} else if kerrors.IsConflict(err) {
						log.Warn("UID of Pod Changed")
					} else {
						log.Err(err).Error("Unknown error while deleting Pod")
					}
				}
			}
		}

		if updateMemberStatusNeeded {
			if err := status.Members.Update(memberStatus, group); err != nil {
				return errors.WithStack(err)
			}
		}

		return nil
	}, podv1.FilterPodsByLabels(k8sutil.LabelsForDeployment(deploymentName, "")))
	if err != nil {
		return 0, err
	}

	// Go over all members, check for missing pods
	for _, e := range status.Members.AsList() {
		m := e.Member
		group := e.Group
		if podName := m.Pod.GetName(); podName != "" {
			if _, exists := cachedStatus.Pod().V1().GetSimple(podName); !exists {
				log.Str("pod-name", podName).Debug("Does not exist")
				switch m.Phase {
				case api.MemberPhaseNone, api.MemberPhasePending, api.MemberPhaseCreationFailed:
					// Do nothing
					log.Str("pod-name", podName).Debug("PodPhase is None, waiting for the pod to be recreated")
				case api.MemberPhaseShuttingDown, api.MemberPhaseUpgrading, api.MemberPhaseFailed, api.MemberPhaseRotateStart, api.MemberPhaseRotating:
					// Shutdown was intended, so not need to do anything here.
					// Just mark terminated
					wasTerminated := m.Conditions.IsTrue(api.ConditionTypeTerminated)
					if m.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Terminated", "") {
						if !wasTerminated {
							// Record termination time
							now := meta.Now()
							m.RecentTerminations = append(m.RecentTerminations, now)
						}
						// Save it
						if err := status.Members.Update(m, group); err != nil {
							return 0, errors.WithStack(err)
						}
					}
				default:
					log.Str("pod-name", podName).Debug("Pod is gone")
					m.Phase = api.MemberPhaseNone // This is trigger a recreate of the pod.
					// Create event
					nextInterval = nextInterval.ReduceTo(recheckSoonPodInspectorInterval)
					events = append(events, k8sutil.NewPodGoneEvent(podName, group.AsRole(), apiObject))
					m.Conditions.Update(api.ConditionTypeReady, false, "Pod Does Not Exist", "")
					wasTerminated := m.Conditions.IsTrue(api.ConditionTypeTerminated)
					if m.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Does Not Exist", "") {
						if !wasTerminated {
							// Record termination time
							now := meta.Now()
							m.RecentTerminations = append(m.RecentTerminations, now)
						}
					}

					// Save it
					if err := status.Members.Update(m, group); err != nil {
						return 0, errors.WithStack(err)
					}
				}
			}
		}
	}

	spec := r.context.GetSpec()
	allMembersReady := status.Members.AllMembersReady(spec.GetMode(), r.context.IsSyncEnabled())
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
	if err := r.context.UpdateStatus(ctx, status); err != nil {
		return 0, errors.WithStack(err)
	}

	// Create events
	for _, evt := range events {
		r.context.CreateEvent(evt)
	}
	return nextInterval, nil
}

func addLabel(labels map[string]string, key, value string) map[string]string {
	if labels != nil {
		labels[key] = value
		return labels
	}

	return map[string]string{
		key: value,
	}
}

func removeLabel(labels map[string]string, key string) map[string]string {
	if labels == nil {
		return map[string]string{}
	}

	delete(labels, key)

	return labels
}

func anyOf(bools ...bool) bool {
	for _, b := range bools {
		if b {
			return true
		}
	}

	return false
}
