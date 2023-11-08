//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package rotation

import (
	"reflect"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/compare"
)

type volumeDiff struct {
	a, b *core.Volume
}

func comparePodVolumes(ds api.DeploymentSpec, g api.ServerGroup, spec, status *core.PodTemplateSpec) compare.Func {
	return func(builder api.ActionBuilder) (mode compare.Mode, plan api.Plan, err error) {
		specV := mapVolumes(spec.Spec)
		statusV := mapVolumes(status.Spec)

		diff := getVolumesDiffFromPods(specV, statusV)

		if len(diff) == 0 {
			return compare.SkippedRotation, nil, nil
		}

		for k, v := range diff {
			switch k {
			case shared.ArangoDTimezoneVolumeName:
				// We are fine, should be just replaced
				if v.a == nil {
					// we remove volume
					return compare.GracefulRotation, nil, nil
				}

				if ds.Mode.Get().ServingGroup() == g {
					// Always enforce on serving group
					return compare.GracefulRotation, nil, nil
				}
			default:
				return compare.GracefulRotation, nil, nil
			}
		}

		status.Spec.Volumes = spec.Spec.Volumes
		return compare.SilentRotation, nil, nil
	}
}

func getVolumesDiffFromPods(a, b map[string]*core.Volume) map[string]volumeDiff {
	d := map[string]volumeDiff{}

	for k := range a {
		if z, ok := b[k]; ok {
			if !reflect.DeepEqual(a[k], z) {
				d[k] = volumeDiff{
					a: a[k],
					b: z,
				}
			}
		} else {
			d[k] = volumeDiff{
				a: a[k],
				b: nil,
			}
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			d[k] = volumeDiff{
				a: nil,
				b: b[k],
			}
		}
	}

	return d
}

func mapVolumes(a core.PodSpec) map[string]*core.Volume {
	n := make(map[string]*core.Volume, len(a.Volumes))

	for id := range a.Volumes {
		v := &a.Volumes[id]

		n[v.Name] = v
	}

	return n
}
