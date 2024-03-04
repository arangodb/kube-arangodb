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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.Pod[Volumes] = &Volumes{}

type Volumes struct {
	// Volumes keeps list of volumes that can be mounted by containers belonging to the pod.
	// +doc/type: []core.Volume
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/concepts/storage/volumes
	Volumes []core.Volume `json:"volumes,omitempty"`
}

func (v *Volumes) Apply(template *core.PodTemplateSpec) error {
	if v == nil {
		return nil
	}

	obj := v.DeepCopy()

	template.Spec.Volumes = obj.Volumes

	return nil
}

func (v *Volumes) With(other *Volumes) *Volumes {
	if v == nil && other == nil {
		return nil
	}

	if v == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return v.DeepCopy()
	}

	return &Volumes{
		Volumes: resources.MergeVolumes(v.Volumes, other.Volumes...),
	}
}

func (v *Volumes) Validate() error {
	return nil
}
