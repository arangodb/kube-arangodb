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
// Author Ewout Prangsma
//

package k8sutil

import v1 "k8s.io/api/core/v1"

// GetContainerByName returns the container in the given pod with the given name.
// Returns false if not found.
func GetContainerByName(p *v1.Pod, name string) (v1.Container, bool) {
	for _, c := range p.Spec.Containers {
		if c.Name == name {
			return c, true
		}
	}
	return v1.Container{}, false
}

// IsResourceRequirementsChanged returns true if the resource requirements have changed.
func IsResourceRequirementsChanged(wanted, given v1.ResourceRequirements) bool {
	checkList := func(wanted, given v1.ResourceList) bool {
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
