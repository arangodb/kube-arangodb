//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
)

func compareServerContainerVolumeMounts(ds api.DeploymentSpec, g api.ServerGroup, spec, status *core.Container) comparePodContainerFunc {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error) {
		specV := mapVolumeMounts(spec)
		statusV := mapVolumeMounts(status)

		diff := getVolumeMountsDiffFromPods(specV, statusV)

		if len(diff) == 0 {
			return SkippedRotation, nil, nil
		}

		for k, v := range diff {
			switch k {
			case shared.ArangoDTimezoneVolumeName:
				// We are fine, should be just replaced
				if v.a == nil {
					// we remove volume
					return GracefulRotation, nil, nil
				}

				if ds.Mode.Get().ServingGroup() == g {
					// Always enforce on serving group
					return GracefulRotation, nil, nil
				}
			case shared.LifecycleVolumeName:
				// Do nothing
			default:
				return GracefulRotation, nil, nil
			}
		}

		status.VolumeMounts = spec.VolumeMounts
		return SilentRotation, nil, nil
	}
}

func compareAnyContainerVolumeMounts(ds api.DeploymentSpec, g api.ServerGroup, spec, status *core.Container) comparePodContainerFunc {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error) {
		specV := mapVolumeMounts(spec)
		statusV := mapVolumeMounts(status)

		diff := getVolumeMountsDiffFromPods(specV, statusV)

		if len(diff) == 0 {
			return SkippedRotation, nil, nil
		}

		for k := range diff {
			switch k {
			case shared.LifecycleVolumeName:
				// Do nothing
			default:
				return GracefulRotation, nil, nil
			}
		}

		status.VolumeMounts = spec.VolumeMounts
		return SilentRotation, nil, nil
	}
}

type volumeMountDiff struct {
	a, b []*core.VolumeMount
}

func getVolumeMountsDiffFromPods(a, b map[string][]*core.VolumeMount) map[string]volumeMountDiff {
	d := map[string]volumeMountDiff{}

	for k := range a {
		if z, ok := b[k]; ok {
			if !reflect.DeepEqual(a[k], z) {
				d[k] = volumeMountDiff{
					a: a[k],
					b: z,
				}
			}
		} else {
			d[k] = volumeMountDiff{
				a: a[k],
				b: nil,
			}
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			d[k] = volumeMountDiff{
				a: nil,
				b: a[k],
			}
		}
	}

	return d
}

func mapVolumeMounts(a *core.Container) map[string][]*core.VolumeMount {
	n := make(map[string][]*core.VolumeMount, len(a.VolumeMounts))

	for id := range a.VolumeMounts {
		v := &a.VolumeMounts[id]

		n[v.Name] = append(n[v.Name], v)
	}

	return n
}
