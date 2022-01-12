//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import core "k8s.io/api/core/v1"

// GetContainerByName returns the container in the given pod with the given name.
// Returns false if not found.
func GetContainerByName(p *core.Pod, name string) (core.Container, bool) {
	for _, c := range p.Spec.Containers {
		if c.Name == name {
			return c, true
		}
	}
	return core.Container{}, false
}

// GetContainerStatusByName returns the container status in the given pod with the given name.
// Returns false if not found.
func GetContainerStatusByName(p *core.Pod, name string) (core.ContainerStatus, bool) {
	for _, c := range p.Status.ContainerStatuses {
		if c.Name == name {
			return c, true
		}
	}
	return core.ContainerStatus{}, false
}

// GetAnyContainerByName returns the container in the given containers with the given name.
// Returns false if not found.
func GetAnyContainerByName(containers []core.Container, name string) (core.Container, bool) {
	for _, c := range containers {
		if c.Name == name {
			return c, true
		}
	}
	return core.Container{}, false
}

// GetAnyContainerStatusByName returns the container status in the given ContainerStatus list with the given name.
// Returns false if not found.
func GetAnyContainerStatusByName(containers []core.ContainerStatus, name string) (core.ContainerStatus, bool) {
	for _, c := range containers {
		if c.Name == name {
			return c, true
		}
	}
	return core.ContainerStatus{}, false
}

// GetFailedContainerNames returns list of failed containers from provided list of statuses.
func GetFailedContainerNames(containers []core.ContainerStatus) []string {
	var failedContainers []string

	for _, c := range containers {
		if IsContainerFailed(&c) {
			failedContainers = append(failedContainers, c.Name)
		}
	}

	return failedContainers
}

// IsResourceRequirementsChanged returns true if the resource requirements have changed.
func IsResourceRequirementsChanged(wanted, given core.ResourceRequirements) bool {
	checkList := func(wanted, given core.ResourceList) bool {
		for k, v := range wanted {
			if gv, ok := given[k]; !ok {
				return true
			} else if v.Cmp(gv) != 0 {
				return true
			}
		}

		return false
	}

	return checkList(wanted.Limits, given.Limits) || checkList(wanted.Requests, given.Requests)
}
