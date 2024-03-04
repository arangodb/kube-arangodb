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

import core "k8s.io/api/core/v1"

func MergeVolumes(in []core.Volume, envs ...core.Volume) []core.Volume {
	out := append([]core.Volume{}, in...)

	for _, env := range envs {
		var envCopy core.Volume
		env.DeepCopyInto(&envCopy)
		if id := VolumeID(out, envCopy.Name); id == -1 {
			out = append(out, envCopy)
		} else {
			out[id] = envCopy
		}
	}

	return out
}

func MergeVolumeMounts(in []core.VolumeMount, envs ...core.VolumeMount) []core.VolumeMount {
	out := append([]core.VolumeMount{}, in...)

	out = append(out, envs...)

	return out
}

func VolumeID(in []core.Volume, name string) int {
	for id := range in {
		if in[id].Name == name {
			return id
		}
	}

	return -1
}
