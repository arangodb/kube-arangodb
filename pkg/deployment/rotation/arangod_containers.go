//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
// Author Tomasz Mielech
//

package rotation

import (
	"strings"

	"github.com/arangodb/kube-arangodb/pkg/deployment/topology"

	"k8s.io/apimachinery/pkg/api/equality"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	core "k8s.io/api/core/v1"
)

const (
	ContainerName  = "name"
	ContainerImage = "image"
)

func containersCompare(_ api.DeploymentSpec, _ api.ServerGroup, spec, status *core.PodSpec) compareFunc {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error) {
		a, b := spec.Containers, status.Containers

		if len(a) == 0 || len(a) != len(b) {
			// If the number of the containers is different or is zero then skip rotation.
			return SkippedRotation, nil, nil
		}

		for id := range a {
			if ac, bc := &a[id], &b[id]; ac.Name == bc.Name {
				if ac.Name == api.ServerGroupReservedContainerNameServer {
					if isOnlyLogLevelChanged(ac.Command, bc.Command) {
						plan = append(plan, builder.NewAction(api.ActionTypeRuntimeContainerArgsLogLevelUpdate).
							AddParam(ContainerName, ac.Name))

						bc.Command = ac.Command
						mode = mode.And(InPlaceRotation)
					}

					if !equality.Semantic.DeepEqual(ac.Env, bc.Env) {
						if areEnvsEqual(ac.Env, bc.Env, func(a, b map[string]core.EnvVar) (map[string]core.EnvVar, map[string]core.EnvVar) {
							if _, ok := a[topology.ArangoDBZone]; !ok {
								delete(b, topology.ArangoDBZone)
							}

							return a, b
						}) {
							bc.Env = ac.Env
							mode = mode.And(SilentRotation)
						}
					}
				} else {
					if ac.Image != bc.Image {
						// Image changed
						plan = append(plan, builder.NewAction(api.ActionTypeRuntimeContainerImageUpdate).AddParam(ContainerName, ac.Name).AddParam(ContainerImage, ac.Image))

						bc.Image = ac.Image
						mode = mode.And(InPlaceRotation)
					}
				}

				if api.IsReservedServerGroupContainerName(ac.Name) {
					mode = mode.And(internalContainerLifecycleCompare(ac, bc))
				}
			}
		}

		return
	}
}

func initContainersCompare(deploymentSpec api.DeploymentSpec, group api.ServerGroup, spec, status *core.PodSpec) compareFunc {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error) {
		gs := deploymentSpec.GetServerGroupSpec(group)

		switch gs.InitContainers.GetMode().Get() {
		case api.ServerGroupInitContainerIgnoreMode:
			// Just copy spec to status if different
			if equal, err := util.CompareJSON(spec.InitContainers, status.InitContainers); err != nil {
				return 0, nil, err
			} else if !equal {
				status.InitContainers = spec.InitContainers
				mode = mode.And(SilentRotation)
			} else {
				return 0, nil, err
			}
		default:
			if len(status.InitContainers) != len(spec.InitContainers) {
				// Nothing to do, count is different
				return
			}

			for id := range status.InitContainers {
				if status.InitContainers[id].Name != spec.InitContainers[id].Name {
					// Nothing to do, order is different
					return
				}
			}

			for id := range status.InitContainers {
				if api.IsReservedServerGroupInitContainerName(status.InitContainers[id].Name) {
					if equal, err := util.CompareJSON(spec.InitContainers[id], status.InitContainers[id]); err != nil {
						return 0, nil, err
					} else if !equal {
						status.InitContainers[id] = spec.InitContainers[id]
						mode = mode.And(SilentRotation)
					}
				}
			}
		}

		return
	}
}

// isOnlyLogLevelChanged returns true when status and spec log level arguments are different.
// If any other argument than --log.level is different false is returned.
func isOnlyLogLevelChanged(specArgs, statusArgs []string) bool {
	diff := util.DiffStrings(specArgs, statusArgs)
	if len(diff) == 0 {
		return false
	}

	for _, arg := range diff {
		if !strings.HasPrefix(strings.TrimLeft(arg, " "), "--log.level") {
			return false
		}
	}

	return true
}

func internalContainerLifecycleCompare(spec, status *core.Container) Mode {
	if spec.Lifecycle == nil && status.Lifecycle == nil {
		return SkippedRotation
	}

	if spec.Lifecycle == nil {
		status.Lifecycle = nil
		return SilentRotation
	}

	if status.Lifecycle == nil {
		status.Lifecycle = spec.Lifecycle
		return SilentRotation
	}

	if !equality.Semantic.DeepEqual(spec.Lifecycle, status.Lifecycle) {
		status.Lifecycle = spec.Lifecycle.DeepCopy()
		return SilentRotation
	}

	return SkippedRotation
}

func areEnvsEqual(a, b []core.EnvVar, rules ...func(a, b map[string]core.EnvVar) (map[string]core.EnvVar, map[string]core.EnvVar)) bool {
	am := getEnvs(a)
	bm := getEnvs(b)

	for _, r := range rules {
		am, bm = r(am, bm)
	}

	return equality.Semantic.DeepEqual(am, bm)
}

func getEnvs(e []core.EnvVar) map[string]core.EnvVar {
	m := map[string]core.EnvVar{}

	for _, q := range e {
		m[q.Name] = q
	}

	return m
}
