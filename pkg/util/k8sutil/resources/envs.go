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

func MergeEnvs(in []core.EnvVar, envs ...core.EnvVar) []core.EnvVar {
	out := append([]core.EnvVar{}, in...)

	for _, env := range envs {
		var envCopy core.EnvVar
		env.DeepCopyInto(&envCopy)
		if id := EnvId(out, envCopy.Name); id == -1 {
			out = append(out, envCopy)
		} else {
			out[id] = envCopy
		}
	}

	return out
}

func MergeEnvFrom(in []core.EnvFromSource, envs ...core.EnvFromSource) []core.EnvFromSource {
	out := append([]core.EnvFromSource{}, in...)

	out = append(out, envs...)

	return out
}

func EnvId(in []core.EnvVar, name string) int {
	for id := range in {
		if in[id].Name == name {
			return id
		}
	}

	return -1
}
