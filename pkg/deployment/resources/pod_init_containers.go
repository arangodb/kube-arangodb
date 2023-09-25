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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// applyInitContainersResourceResources updates passed init containers to ensure that all resources are set (if such feature is enabled)
// This is applied only to containers added by operator itself
func applyInitContainersResourceResources(initContainers []core.Container, mainContainerResources core.ResourceRequirements) []core.Container {
	if !features.InitContainerCopyResources().Enabled() {
		return initContainers
	}

	for i := range initContainers {
		if !api.IsReservedServerGroupInitContainerName(initContainers[i].Name) {
			continue
		}

		k8sutil.ApplyContainerResourceRequirements(&initContainers[i], mainContainerResources)
	}
	return initContainers
}
