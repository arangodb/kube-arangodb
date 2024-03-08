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
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.Container[Lifecycle] = &Lifecycle{}

type Lifecycle struct {
	// Lifecycle keeps actions that the management system should take in response to container lifecycle events.
	// +doc/type: core.Lifecycle
	Lifecycle *core.Lifecycle `json:"lifecycle,omitempty"`
}

func (n *Lifecycle) Apply(_ *core.PodTemplateSpec, template *core.Container) error {
	if n == nil {
		return nil
	}

	template.Lifecycle = n.Lifecycle.DeepCopy()

	return nil
}

func (n *Lifecycle) With(newResources *Lifecycle) *Lifecycle {
	if n == nil && newResources == nil {
		return nil
	}

	if n == nil {
		return newResources.DeepCopy()
	}

	if newResources == nil {
		return n.DeepCopy()
	}

	return &Lifecycle{
		Lifecycle: kresources.MergeLifecycle(n.Lifecycle, newResources.Lifecycle),
	}
}

func (n *Lifecycle) Validate() error {
	return nil
}
