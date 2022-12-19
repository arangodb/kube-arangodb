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
	"strings"

	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	strings2 "github.com/arangodb/kube-arangodb/pkg/util/strings"
)

const (
	ContainerName  = "name"
	ContainerImage = "image"
)

func containersCompare(ds api.DeploymentSpec, g api.ServerGroup, spec, status *core.PodSpec) comparePodFunc {
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

					g := podContainerFuncGenerator(ds, g, ac, bc)

					if m, p, err := comparePodContainer(builder, g(compareServerContainerVolumeMounts), g(compareServerContainerProbes), g(compareServerContainerEnvs)); err != nil {
						log.Err(err).Msg("Error while getting pod diff")
						return SkippedRotation, nil, err
					} else {
						mode = mode.And(m)
						plan = append(plan, p...)
					}

					if !equality.Semantic.DeepEqual(ac.EnvFrom, bc.EnvFrom) {
						// Check EnvFromSource differences.
						filter := func(a, b map[string]core.EnvFromSource) (map[string]core.EnvFromSource, map[string]core.EnvFromSource) {
							delete(a, features.ConfigMapName())
							delete(b, features.ConfigMapName())

							return a, b
						}
						if areEnvsFromEqual(ac.EnvFrom, bc.EnvFrom, filter) {
							// Envs are the same after filtering, but it were different before filtering, so it can be replaced.
							bc.EnvFrom = ac.EnvFrom
							mode = mode.And(SilentRotation)
						}
					}

					if !equality.Semantic.DeepEqual(ac.Ports, bc.Ports) {
						bc.Ports = ac.Ports
						mode = mode.And(SilentRotation)
					}
				} else {
					if ac.Image != bc.Image {
						// Image changed
						plan = append(plan, builder.NewAction(api.ActionTypeRuntimeContainerImageUpdate).AddParam(ContainerName, ac.Name).AddParam(ContainerImage, ac.Image))

						bc.Image = ac.Image
						mode = mode.And(InPlaceRotation)
					}

					g := podContainerFuncGenerator(ds, g, ac, bc)

					if m, p, err := comparePodContainer(builder, g(compareAnyContainerVolumeMounts), g(compareAnyContainerEnvs)); err != nil {
						log.Err(err).Msg("Error while getting pod diff")
						return SkippedRotation, nil, err
					} else {
						mode = mode.And(m)
						plan = append(plan, p...)
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

func initContainersCompare(deploymentSpec api.DeploymentSpec, group api.ServerGroup, spec, status *core.PodSpec) comparePodFunc {
	return func(builder api.ActionBuilder) (Mode, api.Plan, error) {
		gs := deploymentSpec.GetServerGroupSpec(group)

		equal, err := util.CompareJSON(spec.InitContainers, status.InitContainers)
		if err != nil {
			return SkippedRotation, nil, err
		}

		// if equal nothing to do
		if equal {
			return SkippedRotation, nil, nil
		}

		switch gs.InitContainers.GetMode().Get() {
		case api.ServerGroupInitContainerIgnoreMode:
			// Just copy spec to status if different
			if !equal {
				status.InitContainers = spec.InitContainers
				return SilentRotation, nil, err
			} else {
				return SkippedRotation, nil, err
			}
		default:
			statusInitContainers, specInitContainers := filterReservedInitContainers(status.InitContainers), filterReservedInitContainers(spec.InitContainers)
			if equal, err := util.CompareJSON(specInitContainers, statusInitContainers); err != nil {
				return SkippedRotation, nil, err
			} else if equal {
				status.InitContainers = spec.InitContainers
				return SilentRotation, nil, nil
			}
		}

		return SkippedRotation, nil, nil
	}
}

// filterReservedInitContainers filters out reserved container names (which does not enforce restarts)
func filterReservedInitContainers(c []core.Container) []core.Container {
	r := make([]core.Container, 0, len(c))

	for id := range c {
		if api.IsReservedServerGroupInitContainerName(c[id].Name) {
			continue
		}

		r = append(r, c[id])
	}

	return r
}

// isOnlyLogLevelChanged returns true when status and spec log level arguments are different.
// If any other argument than --log.level is different false is returned.
func isOnlyLogLevelChanged(specArgs, statusArgs []string) bool {
	diff := strings2.DiffStrings(specArgs, statusArgs)
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

func areProbesEqual(a, b *core.Probe) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return equality.Semantic.DeepEqual(a, b)
}

func isManagedProbe(a, b *core.Probe) bool {
	if a.Exec == nil || b.Exec == nil {
		return false
	}

	if len(a.Exec.Command) == 0 || len(b.Exec.Command) == 0 {
		return false
	}

	return a.Exec.Command[0] == k8sutil.LifecycleBinary() && b.Exec.Command[0] == k8sutil.LifecycleBinary()
}

// areEnvsFromEqual returns true when environment variables from source are the same after filtering.
func areEnvsFromEqual(a, b []core.EnvFromSource, rules ...func(a, b map[string]core.EnvFromSource) (map[string]core.EnvFromSource, map[string]core.EnvFromSource)) bool {
	am := createEnvsFromMap(a)
	bm := createEnvsFromMap(b)

	for _, r := range rules {
		am, bm = r(am, bm)
	}

	return equality.Semantic.DeepEqual(am, bm)
}

// createEnvsFromMap returns map from list.
func createEnvsFromMap(e []core.EnvFromSource) map[string]core.EnvFromSource {
	m := map[string]core.EnvFromSource{}

	for _, q := range e {
		if q.ConfigMapRef != nil {
			m[q.ConfigMapRef.Name] = q
		} else if q.SecretRef != nil {
			m[q.SecretRef.Name] = q
		}
	}

	return m
}
