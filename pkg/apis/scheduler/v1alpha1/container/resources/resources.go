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

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/interfaces"
	"github.com/arangodb/kube-arangodb/pkg/util"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.Container[Resources] = &Resources{}

type Resources struct {
	// Resources holds resource requests & limits for container
	// +doc/type: core.ResourceRequirements
	// +doc/link: Documentation of core.ResourceRequirements|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core
	Resources *core.ResourceRequirements `json:"resources,omitempty"`
}

func (r *Resources) Apply(_ *core.PodTemplateSpec, template *core.Container) error {
	if r == nil {
		return nil
	}

	template.Resources = kresources.CleanContainerResource(kresources.MergeContainerResource(template.Resources, r.GetResources()))

	return nil
}

func (r *Resources) With(newResources *Resources) *Resources {
	if r == nil && newResources == nil {
		return nil
	}

	if r == nil {
		return newResources.DeepCopy()
	}

	if newResources == nil {
		return r.DeepCopy()
	}

	return &Resources{Resources: util.NewType(kresources.MergeContainerResource(r.GetResources(), newResources.GetResources()))}
}

func (r *Resources) GetResources() core.ResourceRequirements {
	if r == nil || r.Resources == nil {
		return core.ResourceRequirements{}
	}

	local := r.Resources.DeepCopy()

	local.Limits = kresources.UpscaleOptionalContainerResourceList(local.Limits, local.Requests)

	return *local
}

func (r *Resources) Validate() error {
	return nil
}
