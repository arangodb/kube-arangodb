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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/api/core/v1"
)

// createPlan considers the current specification & status of the deployment creates a plan to
// get the status in line with the specification.
// If a plan already exists, nothing is done.
func (d *Deployment) createPlan() error {
	// Get all current pods
	pods, err := d.deps.KubeCli.CoreV1().Pods(d.apiObject.GetNamespace()).List(k8sutil.DeploymentListOpt(d.apiObject.GetName()))
	if err != nil {
		log.Debug().Err(err).Msg("Failed to list pods")
		return maskAny(err)
	}
	myPods := make([]v1.Pod, 0, len(pods.Items))
	for _, p := range pods.Items {
		if d.isOwnerOf(&p) {
			myPods = append(myPods, p)
		}
	}

	// Create plan
	newPlan, changed := createPlan(d.deps.Log, d.status.Plan, d.apiObject.Spec, d.status, myPods)

	// If not change, we're done
	if !changed {
		return nil
	}

	// Save plan
	if len(newPlan) == 0 {
		// Nothing to do
		return nil
	}
	d.status.Plan = newPlan
	if err := d.updateCRStatus(); err != nil {
		return maskAny(err)
	}
	return nil
}

// createPlan considers the given specification & status and creates a plan to get the status in line with the specification.
// If a plan already exists, the given plan is returned with false.
// Otherwise the new plan is returned with a boolean true.
func createPlan(log zerolog.Logger, currentPlan api.Plan, spec api.DeploymentSpec, status api.DeploymentStatus, pods []v1.Pod) (api.Plan, bool) {
	if len(currentPlan) > 0 {
		// Plan already exists, complete that first
		return currentPlan, false
	}

	// Check for various scenario's
	var plan api.Plan

	// Check for scale up/down
	switch spec.Mode {
	case api.DeploymentModeSingle:
		// Never scale down
	case api.DeploymentModeResilientSingle:
		// Only scale singles
		plan = append(plan, createScalePlan(log, status.Members.Single, api.ServerGroupSingle, spec.Single.Count)...)
	case api.DeploymentModeCluster:
		// Scale dbservers, coordinators, syncmasters & syncworkers
		plan = append(plan, createScalePlan(log, status.Members.DBServers, api.ServerGroupDBServers, spec.DBServers.Count)...)
		plan = append(plan, createScalePlan(log, status.Members.Coordinators, api.ServerGroupCoordinators, spec.Coordinators.Count)...)
		plan = append(plan, createScalePlan(log, status.Members.SyncMasters, api.ServerGroupSyncMasters, spec.SyncMasters.Count)...)
		plan = append(plan, createScalePlan(log, status.Members.SyncWorkers, api.ServerGroupSyncWorkers, spec.SyncWorkers.Count)...)
	}

	// Check for the need to rotate one or more members
	getPod := func(podName string) *v1.Pod {
		for _, p := range pods {
			if p.GetName() == podName {
				return &p
			}
		}
		return nil
	}
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members *api.MemberStatusList) error {
		for _, m := range *members {
			if len(plan) > 0 {
				// Only 1 change at a time
				continue
			}
			if m.State != api.MemberStateCreated {
				// Only rotate when state is created
				continue
			}
			if podName := m.PodName; podName != "" {
				if p := getPod(podName); p != nil {
					// Got pod, compare it with what it should be
					if podNeedsRotation(*p, spec) {
						plan = append(plan, createRotateMemberPlan(log, m, group)...)
					}
				}
			}
		}
		return nil
	})

	// Return plan
	return plan, true
}

// podNeedsRotation returns true when the specification of the
// given pod differs from what it should be according to the
// given deployment spec.
func podNeedsRotation(p v1.Pod, spec api.DeploymentSpec) bool {
	// Check number of containers
	if len(p.Spec.Containers) != 1 {
		return true
	}
	// Check image
	c := p.Spec.Containers[0]
	if c.Image != spec.Image || c.ImagePullPolicy != spec.ImagePullPolicy {
		return true
	}
	// Check arguments
	// TODO

	return false
}

// createScalePlan creates a scaling plan for a single server group
func createScalePlan(log zerolog.Logger, members api.MemberStatusList, group api.ServerGroup, count int) api.Plan {
	var plan api.Plan
	if len(members) < count {
		// Scale up
		toAdd := count - len(members)
		for i := 0; i < toAdd; i++ {
			plan = append(plan, api.NewAction(api.ActionTypeAddMember, group, ""))
		}
		log.Debug().
			Int("delta", toAdd).
			Str("role", group.AsRole()).
			Msg("Creating scale-up plan")
	} else if len(members) > count {
		// Note, we scale down 1 member as a time
		if m, err := members.SelectMemberToRemove(); err == nil {
			if group == api.ServerGroupDBServers {
				plan = append(plan,
					api.NewAction(api.ActionTypeCleanOutMember, group, m.ID),
				)
			}
			plan = append(plan,
				api.NewAction(api.ActionTypeShutdownMember, group, m.ID),
				api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
			)
			log.Debug().
				Str("role", group.AsRole()).
				Msg("Creating scale-down plan")
		}
	}
	return plan
}

// createRotateMemberPlan creates a plan to rotate (stop-recreate-start) an existing
// member.
func createRotateMemberPlan(log zerolog.Logger, member api.MemberStatus, group api.ServerGroup) api.Plan {
	log.Debug().
		Str("id", member.ID).
		Str("role", group.AsRole()).
		Msg("Creating rotation plan")
	plan := api.Plan{
		api.NewAction(api.ActionTypeRotateMember, group, member.ID),
		api.NewAction(api.ActionTypeWaitForMemberUp, group, member.ID),
	}
	return plan
}
