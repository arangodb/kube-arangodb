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

package v1

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

type Resources struct {
	// Resources holds resource requests & limits for container
	// +doc/type: core.ResourceRequirements
	// +doc/link: Documentation of core.ResourceRequirements|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core
	Resources *core.ResourceRequirements `json:"resources,omitempty"`
}

func (r *Resources) With(newResources core.ResourceRequirements) core.ResourceRequirements {
	if res := r.GetResources(); res == nil {
		return newResources
	} else {
		return resources.ApplyContainerResource(*res, newResources)
	}
}

func (r *Resources) GetResources() *core.ResourceRequirements {
	if r == nil || r.Resources == nil {
		return nil
	}

	return r.Resources
}

func (r *Resources) Validate() error {
	return nil
}
