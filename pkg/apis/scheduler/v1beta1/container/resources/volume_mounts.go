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

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/interfaces"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.Container[VolumeMounts] = &VolumeMounts{}

type VolumeMounts struct {
	// VolumeMounts keeps list of pod volumes to mount into the container's filesystem.
	// +doc/type: []core.VolumeMount
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`
}

func (v *VolumeMounts) Apply(_ *core.PodTemplateSpec, container *core.Container) error {
	if v == nil {
		return nil
	}

	obj := v.DeepCopy()

	container.VolumeMounts = kresources.MergeVolumeMounts(container.VolumeMounts, obj.VolumeMounts...)

	return nil
}

func (v *VolumeMounts) With(other *VolumeMounts) *VolumeMounts {
	if v == nil && other == nil {
		return nil
	}

	if v == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return v.DeepCopy()
	}

	return &VolumeMounts{
		VolumeMounts: kresources.MergeVolumeMounts(v.VolumeMounts, other.VolumeMounts...),
	}
}

func (v *VolumeMounts) Validate() error {
	return nil
}
