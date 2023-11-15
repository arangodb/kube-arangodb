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
	"strings"

	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/compare"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	arangoStrings "github.com/arangodb/kube-arangodb/pkg/util/strings"
)

const (
	ContainerName  = "name"
	ContainerImage = "image"
)

func containersCompare(ds api.DeploymentSpec, g api.ServerGroup, spec, status *core.PodTemplateSpec) compare.Func {
	return compare.SubElementsP2(func(in *core.PodTemplateSpec) *[]core.Container {
		return &in.Spec.Containers
	}, func(ds api.DeploymentSpec, g api.ServerGroup, specContainers, statusContainers *[]core.Container) compare.Func {
		return compare.ArrayExtractorP2(func(ds api.DeploymentSpec, g api.ServerGroup, specContainer, statusContainer *core.Container) compare.Func {
			return func(builder api.ActionBuilder) (mode compare.Mode, plan api.Plan, err error) {
				if specContainer.Name != statusContainer.Name {
					return compare.SkippedRotation, nil, nil
				}

				if specContainer.Name == api.ServerGroupReservedContainerNameServer {
					if isOnlyLogLevelChanged(specContainer.Command, statusContainer.Command) {
						plan = append(plan, builder.NewAction(api.ActionTypeRuntimeContainerArgsLogLevelUpdate).
							AddParam(ContainerName, specContainer.Name))

						statusContainer.Command = specContainer.Command
						mode = mode.And(compare.InPlaceRotation)
					}

					g := compare.NewFuncGenP2(ds, g, specContainer, statusContainer)

					if m, p, err := compare.Evaluate(builder, g(compareServerContainerVolumeMounts), g(compareServerContainerProbes), g(compareServerContainerEnvs)); err != nil {
						log.Err(err).Msg("Error while getting pod diff")
						return compare.SkippedRotation, nil, err
					} else {
						mode = mode.And(m)
						plan = append(plan, p...)
					}

					if !equality.Semantic.DeepEqual(specContainer.EnvFrom, statusContainer.EnvFrom) {
						// Check EnvFromSource differences.
						filter := func(a, b map[string]core.EnvFromSource) (map[string]core.EnvFromSource, map[string]core.EnvFromSource) {
							delete(a, features.ConfigMapName())
							delete(b, features.ConfigMapName())

							return a, b
						}
						if areEnvsFromEqual(specContainer.EnvFrom, statusContainer.EnvFrom, filter) {
							// Envs are the same after filtering, but it were different before filtering, so it can be replaced.
							statusContainer.EnvFrom = specContainer.EnvFrom
							mode = mode.And(compare.SilentRotation)
						}
					}

					if !equality.Semantic.DeepEqual(specContainer.Ports, statusContainer.Ports) {
						statusContainer.Ports = specContainer.Ports
						mode = mode.And(compare.SilentRotation)
					}
				} else {
					if specContainer.Image != statusContainer.Image {
						// Image changed
						plan = append(plan, builder.NewAction(api.ActionTypeRuntimeContainerImageUpdate).AddParam(ContainerName, specContainer.Name).AddParam(ContainerImage, specContainer.Image))

						statusContainer.Image = specContainer.Image
						mode = mode.And(compare.InPlaceRotation)
					}

					g := compare.NewFuncGenP2(ds, g, specContainer, statusContainer)

					if m, p, err := compare.Evaluate(builder, g(compareAnyContainerVolumeMounts), g(compareAnyContainerEnvs)); err != nil {
						log.Err(err).Msg("Error while getting pod diff")
						return compare.SkippedRotation, nil, err
					} else {
						mode = mode.And(m)
						plan = append(plan, p...)
					}
				}

				if api.IsReservedServerGroupContainerName(specContainer.Name) {
					mode = mode.And(internalContainerLifecycleCompare(specContainer, statusContainer))
				}

				return
			}
		})(ds, g, specContainers, statusContainers)
	})(ds, g, spec, status)
}

func initContainersCompare(deploymentSpec api.DeploymentSpec, group api.ServerGroup, spec, status *core.PodTemplateSpec) compare.Func {
	return func(builder api.ActionBuilder) (compare.Mode, api.Plan, error) {
		gs := deploymentSpec.GetServerGroupSpec(group)

		equal, err := util.CompareJSON(spec.Spec.InitContainers, status.Spec.InitContainers)
		if err != nil {
			return compare.SkippedRotation, nil, err
		}

		// if equal nothing to do
		if equal {
			return compare.SkippedRotation, nil, nil
		}

		switch gs.InitContainers.GetMode().Get() {
		case api.ServerGroupInitContainerIgnoreMode:
			// Just copy spec to status if different
			if !equal {
				status.Spec.InitContainers = spec.Spec.InitContainers
				return compare.SilentRotation, nil, err
			} else {
				return compare.SkippedRotation, nil, err
			}
		default:
			statusInitContainers, specInitContainers := filterReservedInitContainers(status.Spec.InitContainers), filterReservedInitContainers(spec.Spec.InitContainers)
			if equal, err := util.CompareJSON(specInitContainers, statusInitContainers); err != nil {
				return compare.SkippedRotation, nil, err
			} else if equal {
				status.Spec.InitContainers = spec.Spec.InitContainers
				return compare.SilentRotation, nil, nil
			}
		}

		return compare.SkippedRotation, nil, nil
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
	diff := arangoStrings.DiffStrings(specArgs, statusArgs)
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

func internalContainerLifecycleCompare(spec, status *core.Container) compare.Mode {
	if spec.Lifecycle == nil && status.Lifecycle == nil {
		return compare.SkippedRotation
	}

	if spec.Lifecycle == nil {
		status.Lifecycle = nil
		return compare.SilentRotation
	}

	if status.Lifecycle == nil {
		status.Lifecycle = spec.Lifecycle
		return compare.SilentRotation
	}

	if !equality.Semantic.DeepEqual(spec.Lifecycle, status.Lifecycle) {
		status.Lifecycle = spec.Lifecycle.DeepCopy()
		return compare.SilentRotation
	}

	return compare.SkippedRotation
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
