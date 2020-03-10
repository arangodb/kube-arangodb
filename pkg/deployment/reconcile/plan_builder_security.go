//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package reconcile

import (
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createRotateServerStoragePlan creates plan to rotate a server and its volume because of a
// different storage class or a difference in storage resource requirements.
func createRotateServerSecurityPlan(log zerolog.Logger, spec api.DeploymentSpec, status api.DeploymentStatus,
	pods []core.Pod) api.Plan {
	var plan api.Plan
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if !plan.IsEmpty() {
				// Only 1 change at a time
				continue
			}

			groupSpec := spec.GetServerGroupSpec(group)

			pod, found := k8sutil.GetPodByName(pods, m.PodName)
			if !found {
				continue
			}

			container, ok := getServerContainer(pod.Spec.Containers)
			if !ok {
				// We do not have server container in pod, which is not desired
				continue
			}

			groupSC := groupSpec.SecurityContext.NewSecurityContext()
			containerSC := container.SecurityContext

			if !compareSC(groupSC, containerSC) {
				log.Info().Str("member", m.ID).Str("group", group.AsRole()).Msg("Rotating security context")
				plan = append(plan,
					api.NewAction(api.ActionTypeRotateMember, group, m.ID),
					api.NewAction(api.ActionTypeWaitForMemberUp, group, m.ID),
				)
			}
		}
		return nil
	})
	return plan
}

func getServerContainer(containers []core.Container) (core.Container, bool) {
	for _, container := range containers {
		if container.Name == k8sutil.ServerContainerName {
			return container, true
		}
	}

	return core.Container{}, false
}

func compareSC(a, b *core.SecurityContext) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if ok := compareCapabilities(a.Capabilities, b.Capabilities); !ok {
		return false
	}

	return true
}

func compareCapabilities(a, b *core.Capabilities) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if ok := compareCapabilityLists(a.Add, b.Add); !ok {
		return false
	}

	if ok := compareCapabilityLists(a.Drop, b.Drop); !ok {
		return false
	}

	return true
}

func compareCapabilityLists(a, b []core.Capability) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	checked := map[core.Capability]bool{}

	for _, capability := range a {
		checked[capability] = false
	}

	for _, capability := range b {
		if _, ok := checked[capability]; !ok {
			return false
		}

		checked[capability] = true
	}

	for _, check := range checked {
		if !check {
			return false
		}
	}

	return true
}
