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

package deployment

import (
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
)

var (
	inspectedPodCounter = metrics.MustRegisterCounter("deployment", "inspected_pods", "Number of pod inspections")
)

// inspectPods lists all pods that belong to the given deployment and updates
// the member status of the deployment accordingly.
func (d *Deployment) inspectPods() error {
	log := d.deps.Log
	var events []*v1.Event

	pods, err := d.deps.KubeCli.CoreV1().Pods(d.apiObject.GetNamespace()).List(k8sutil.DeploymentListOpt(d.apiObject.GetName()))
	if err != nil {
		log.Debug().Err(err).Msg("Failed to list pods")
		return maskAny(err)
	}

	// Update member status from all pods found
	for _, p := range pods.Items {
		// Check ownership
		if !d.isOwnerOf(&p) {
			log.Debug().Str("pod", p.GetName()).Msg("pod not owned by this deployment")
			continue
		}

		// Pod belongs to this deployment, update metric
		inspectedPodCounter.Inc()

		// Find member status
		memberStatus, group, found := d.status.Members.MemberStatusByPodName(p.GetName())
		if !found {
			log.Debug().Str("pod", p.GetName()).Msg("no memberstatus found for pod")
			continue
		}

		// Update state
		updateMemberStatusNeeded := false
		if k8sutil.IsPodSucceeded(&p) {
			// Pod has terminated with exit code 0.
			if memberStatus.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Succeeded", "") {
				updateMemberStatusNeeded = true
			}
		} else if k8sutil.IsPodFailed(&p) {
			// Pod has terminated with at least 1 container with a non-zero exit code.
			if memberStatus.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Failed", "") {
				updateMemberStatusNeeded = true
			}
		}
		if k8sutil.IsPodReady(&p) {
			// Pod is now ready
			if memberStatus.Conditions.Update(api.ConditionTypeReady, true, "Pod Ready", "") {
				updateMemberStatusNeeded = true
			}
		} else {
			// Pod is not ready
			if memberStatus.Conditions.Update(api.ConditionTypeReady, false, "Pod Not Ready", "") {
				updateMemberStatusNeeded = true
			}
		}
		if updateMemberStatusNeeded {
			log.Debug().Str("pod-name", p.GetName()).Msg("Updated member status member for pod")
			if err := d.status.Members.UpdateMemberStatus(memberStatus, group); err != nil {
				return maskAny(err)
			}
		}
	}

	podExists := func(podName string) bool {
		for _, p := range pods.Items {
			if p.GetName() == podName && d.isOwnerOf(&p) {
				return true
			}
		}
		return false
	}

	// Go over all members, check for missing pods
	d.status.Members.ForeachServerGroup(func(group api.ServerGroup, members *api.MemberStatusList) error {
		for _, m := range *members {
			if podName := m.PodName; podName != "" {
				if !podExists(podName) {
					switch m.State {
					case api.MemberStateNone:
						// Do nothing
					case api.MemberStateShuttingDown:
						// Shutdown was intended, so not need to do anything here.
						// Just mark terminated
						if m.Conditions.Update(api.ConditionTypeTerminated, true, "Pod Terminated", "") {
							if err := d.status.Members.UpdateMemberStatus(m, group); err != nil {
								return maskAny(err)
							}
						}
					default:
						m.State = api.MemberStateNone // This is trigger a recreate of the pod.
						// Create event
						events = append(events, k8sutil.NewPodGoneEvent(podName, group.AsRole(), d.apiObject))
						if m.Conditions.Update(api.ConditionTypeReady, false, "Pod Does Not Exist", "") {
							if err := d.status.Members.UpdateMemberStatus(m, group); err != nil {
								return maskAny(err)
							}
						}
					}
				}
			}
		}
		return nil
	})

	// Check overall status update
	switch d.status.State {
	case api.DeploymentStateCreating:
		if d.status.Members.AllMembersReady() {
			d.status.State = api.DeploymentStateRunning
		}
		// TODO handle other State values
	}

	// Save status
	if err := d.updateCRStatus(); err != nil {
		return maskAny(err)
	}

	// Create events
	for _, evt := range events {
		d.createEvent(evt)
	}
	return nil
}
