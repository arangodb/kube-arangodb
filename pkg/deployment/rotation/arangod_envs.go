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
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/deployment/topology"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

func compareServerContainerEnvs(ds api.DeploymentSpec, g api.ServerGroup, spec, status *core.Container) comparePodContainerFunc {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error) {
		specV := mapEnvs(spec)
		statusV := mapEnvs(status)

		diff := getEnvDiffFromPods(specV, statusV)

		if len(diff) == 0 {
			return SkippedRotation, nil, nil
		}

		for k := range diff {
			switch k {
			case topology.ArangoDBZone, resources.ArangoDBOverrideServerGroupEnv,
				resources.ArangoDBOverrideDeploymentModeEnv, resources.ArangoDBOverrideVersionEnv,
				resources.ArangoDBOverrideEnterpriseEnv:
				// Those envs can change without restart
				continue
			case constants.EnvOperatorPodName, constants.EnvOperatorPodNamespace, constants.EnvOperatorNodeName, constants.EnvOperatorNodeNameArango:
				// Lifecycle envs can change without restart
				continue
			default:
				return GracefulRotation, nil, nil
			}
		}

		status.Env = spec.Env
		return SilentRotation, nil, nil
	}
}

func compareAnyContainerEnvs(ds api.DeploymentSpec, g api.ServerGroup, spec, status *core.Container) comparePodContainerFunc {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error) {
		specV := mapEnvs(spec)
		statusV := mapEnvs(status)

		diff := getEnvDiffFromPods(specV, statusV)

		if len(diff) == 0 {
			return SkippedRotation, nil, nil
		}

		for k := range diff {
			switch k {
			case constants.EnvOperatorPodName, constants.EnvOperatorPodNamespace, constants.EnvOperatorNodeName, constants.EnvOperatorNodeNameArango:
				// Lifecycle envs can change without restart
				continue
			default:
				return GracefulRotation, nil, nil
			}
		}

		status.Env = spec.Env
		return SilentRotation, nil, nil
	}
}

type envDiff struct {
	a, b []*core.EnvVar
}

func getEnvDiffFromPods(a, b map[string][]*core.EnvVar) map[string]envDiff {
	d := map[string]envDiff{}

	for k := range a {
		if z, ok := b[k]; ok {
			if !reflect.DeepEqual(a[k], z) {
				d[k] = envDiff{
					a: a[k],
					b: z,
				}
			}
		} else {
			d[k] = envDiff{
				a: a[k],
				b: nil,
			}
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			d[k] = envDiff{
				a: nil,
				b: a[k],
			}
		}
	}

	return d
}

func mapEnvs(a *core.Container) map[string][]*core.EnvVar {
	n := make(map[string][]*core.EnvVar, len(a.VolumeMounts))

	for id := range a.Env {
		v := &a.Env[id]

		n[v.Name] = append(n[v.Name], v)
	}

	return n
}
