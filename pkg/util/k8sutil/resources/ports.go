//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

import core "k8s.io/api/core/v1"

func MergeContainerPorts(in []core.ContainerPort, envs ...core.ContainerPort) []core.ContainerPort {
	out := append([]core.ContainerPort{}, in...)

	for _, env := range envs {
		var envCopy core.ContainerPort
		env.DeepCopyInto(&envCopy)
		if id := ContainerPortId(out, envCopy.Name); id == -1 {
			out = append(out, envCopy)
		} else {
			out[id] = envCopy
		}
	}

	return out
}

func ContainerPortId(in []core.ContainerPort, name string) int {
	for id := range in {
		if in[id].Name == name {
			return id
		}
	}

	return -1
}
