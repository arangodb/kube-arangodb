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

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ArangoMemberSpecOverrides struct {
	// VolumeClaimTemplate specifies a template for volume claims. Overrides template provided on the group level.
	// +doc/type: core.PersistentVolumeClaim
	// +doc/link: Documentation of core.PersistentVolumeClaim|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core
	VolumeClaimTemplate *core.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`

	// Resources holds resource requests & limits. Overrides template provided on the group level.
	// +doc/type: core.ResourceRequirements
	// +doc/link: Documentation of core.ResourceRequirements|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core
	Resources core.ResourceRequirements `json:"resources,omitempty"`
}

func (a *ArangoMemberSpecOverrides) HasVolumeClaimTemplate(g *ServerGroupSpec) bool {
	if g != nil {
		if g.HasVolumeClaimTemplate() {
			return true
		}
	}

	if a != nil {
		return a.VolumeClaimTemplate != nil
	}

	return false
}

func (a *ArangoMemberSpecOverrides) GetVolumeClaimTemplate(g *ServerGroupSpec) *core.PersistentVolumeClaim {
	if a != nil {
		if a.VolumeClaimTemplate != nil {
			return a.VolumeClaimTemplate.DeepCopy()
		}
	}

	if g != nil {
		if g.VolumeClaimTemplate != nil {
			return g.VolumeClaimTemplate.DeepCopy()
		}
	}

	return nil
}

func (a *ArangoMemberSpecOverrides) GetResources(g *ServerGroupSpec) core.ResourceRequirements {
	rl := core.ResourceList{}
	rr := core.ResourceList{}

	if g != nil {
		if l := g.Resources.Limits; len(l) > 0 {
			for k, v := range l {
				rl[k] = v
			}
		}
		if l := g.Resources.Requests; len(l) > 0 {
			for k, v := range l {
				rr[k] = v
			}
		}
	}

	if a != nil {
		if l := a.Resources.Limits; len(l) > 0 {
			for k, v := range l {
				rl[k] = v
			}
		}
		if l := a.Resources.Requests; len(l) > 0 {
			for k, v := range l {
				rr[k] = v
			}
		}
	}

	ret := core.ResourceRequirements{}

	if len(rl) > 0 {
		ret.Limits = rl
	}
	if len(rr) > 0 {
		ret.Requests = rr
	}

	return ret
}

func (a *ArangoMemberSpecOverrides) GetStorageClassName(g *ServerGroupSpec) string {
	if a != nil {
		if z := a.VolumeClaimTemplate; z != nil {
			return util.TypeOrDefault[string](z.Spec.StorageClassName)
		}
	}

	if g != nil {
		return g.GetStorageClassName()
	}

	return ""
}
