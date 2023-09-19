//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

import (
	core "k8s.io/api/core/v1"
)

// EnsureAllResourcesNotEmpty copies resource specifications from src to dst if such resource is not defined in dst
func EnsureAllResourcesNotEmpty(src core.ResourceList, dst *core.ResourceList) {
	if dst == nil {
		l := make(core.ResourceList)
		dst = &l
	}
	for k, v := range src {
		if _, ok := (*dst)[k]; !ok {
			(*dst)[k] = v.DeepCopy()
		}
	}
}
