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

package resources

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

// applyInitContainersResourceLimits updates all passed init containers to ensure that resource limits are set (if such feature is enabled)
func applyInitContainersResourceLimits(initContainers []core.Container, mainContainerResources *core.ResourceRequirements) []core.Container {
	if !features.InitContainerCopyLimits().Enabled() || mainContainerResources == nil || len(mainContainerResources.Limits) == 0 {
		return initContainers
	}

	for i, c := range initContainers {
		if len(c.Resources.Limits) == 0 {
			initContainers[i].Resources.Limits = mainContainerResources.Limits.DeepCopy()
		}
	}
	return initContainers
}
