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

package resources

import (
	core "k8s.io/api/core/v1"
)

func ApplyContainerResourceRequirements(container *core.Container, resources core.ResourceRequirements) {
	if container == nil {
		return
	}

	container.Resources.Limits = ApplyContainerResourceList(container.Resources.Limits, resources.Limits)
	container.Resources.Requests = ApplyContainerResourceList(container.Resources.Requests, resources.Requests)
}

// MergeContainerResource updates resources from `from` to `to` ResourceList
func MergeContainerResource(to core.ResourceRequirements, from core.ResourceRequirements) core.ResourceRequirements {
	var r core.ResourceRequirements

	r.Limits = MergeContainerResourceList(to.Limits, from.Limits)
	r.Requests = MergeContainerResourceList(to.Requests, from.Requests)

	return r
}

// ApplyContainerResource adds non-existing resources from `from` to `to` ResourceList
func ApplyContainerResource(to core.ResourceRequirements, from core.ResourceRequirements) core.ResourceRequirements {
	var r core.ResourceRequirements

	r.Limits = ApplyContainerResourceList(to.Limits, from.Limits)
	r.Requests = ApplyContainerResourceList(to.Requests, from.Requests)

	return r
}

// MergeContainerResourceList updates resources from `from` to `to` ResourceList
func MergeContainerResourceList(to core.ResourceList, from core.ResourceList) core.ResourceList {
	if len(from) == 0 {
		return to
	}

	if to == nil {
		to = core.ResourceList{}
	}

	for k, v := range from {
		if v.IsZero() {
			delete(to, k)
		} else {
			to[k] = v
		}
	}

	return to
}

// ApplyContainerResourceList adds non-existing resources from `from` to `to` ResourceList
func ApplyContainerResourceList(to core.ResourceList, from core.ResourceList) core.ResourceList {
	if len(from) == 0 {
		return to
	}

	if to == nil {
		to = core.ResourceList{}
	}

	for k, v := range from {
		if _, ok := to[k]; !ok {
			to[k] = v
		}
	}

	return to
}

func UpscaleContainerResourceRequirements(container *core.Container, resources core.ResourceRequirements) {
	if container == nil {
		return
	}

	container.Resources.Limits = UpscaleContainerResourceList(container.Resources.Limits, resources.Limits)
	container.Resources.Requests = UpscaleContainerResourceList(container.Resources.Requests, resources.Requests)

	// Ensure that Limits are always higher or equals requests
	container.Resources.Limits = UpscaleOptionalContainerResourceList(container.Resources.Limits, container.Resources.Requests)
}

// UpscaleOptionalContainerResourceList scales up resources from `from` to `to` ResourceList if they exists in `to`
func UpscaleOptionalContainerResourceList(to core.ResourceList, from core.ResourceList) core.ResourceList {
	if len(from) == 0 {
		return to
	}

	if to == nil {
		to = core.ResourceList{}
	}

	for k, v := range from {
		if n, ok := to[k]; ok {
			if n.Cmp(v) < 0 {
				to[k] = v
			}
		}
	}

	return to
}

// UpscaleContainerResourceList scales up resources from `from` to `to` ResourceList
func UpscaleContainerResourceList(to core.ResourceList, from core.ResourceList) core.ResourceList {
	if len(from) == 0 {
		return to
	}

	if to == nil {
		to = core.ResourceList{}
	}

	for k, v := range from {
		if n, ok := to[k]; !ok || n.Cmp(v) < 0 {
			to[k] = v
		}
	}

	return to
}

// ExtractPodInitContainerAcceptedResourceRequirement filters resource requirements for InitContainers.
func ExtractPodInitContainerAcceptedResourceRequirement(resources core.ResourceRequirements) core.ResourceRequirements {
	return NewPodResourceRequirementsFilter(PodResourceRequirementsInitContainersAcceptedResourceRequirements()...)(resources)
}

// PodResourceRequirementsInitContainersAcceptedResourceRequirements returns struct if accepted Pod resource types
func PodResourceRequirementsInitContainersAcceptedResourceRequirements() []core.ResourceName {
	return []core.ResourceName{core.ResourceCPU, core.ResourceMemory, core.ResourceEphemeralStorage}
}

// ExtractPodAcceptedResourceRequirement filters resource requirements for Pods.
func ExtractPodAcceptedResourceRequirement(resources core.ResourceRequirements) core.ResourceRequirements {
	return NewPodResourceRequirementsFilter(PodResourceRequirementsPodAcceptedResourceRequirements()...)(resources)
}

// PodResourceRequirementsPodAcceptedResourceRequirements returns struct if accepted Pod resource types
func PodResourceRequirementsPodAcceptedResourceRequirements() []core.ResourceName {
	return []core.ResourceName{core.ResourceCPU, core.ResourceMemory, core.ResourceEphemeralStorage}
}

type PodResourceRequirementsFilter func(in core.ResourceRequirements) core.ResourceRequirements

// NewPodResourceRequirementsFilter returns function which filter out not accepted resources from resource requirements
func NewPodResourceRequirementsFilter(filters ...core.ResourceName) PodResourceRequirementsFilter {
	return func(in core.ResourceRequirements) core.ResourceRequirements {
		filter := NewPodResourceListFilter(filters...)

		return core.ResourceRequirements{
			Limits:   filter(in.Limits),
			Requests: filter(in.Requests),
		}
	}
}

type PodResourceListFilter func(in core.ResourceList) core.ResourceList

// NewPodResourceListFilter returns function which filter out not accepted resources from list
func NewPodResourceListFilter(filters ...core.ResourceName) PodResourceListFilter {
	return func(in core.ResourceList) core.ResourceList {
		filtered := map[core.ResourceName]bool{}

		for _, k := range filters {
			filtered[k] = true
		}

		n := core.ResourceList{}

		for k, v := range in {
			if _, ok := filtered[k]; !ok {
				continue
			}

			n[k] = v
		}

		return n
	}
}
