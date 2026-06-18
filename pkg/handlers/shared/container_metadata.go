//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	"sort"

	core "k8s.io/api/core/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func AddMetadataToContainers(secret *sharedApi.Object, envs map[string]string, containers ...string) (*schedulerApi.ProfileTemplate, error) {
	cenvs := []core.EnvVar{
		{
			Name:  "SECRET_METADATA_CHECKSUM",
			Value: secret.GetChecksum(),
		},
	}
	for k, v := range envs {
		cenvs = append(cenvs, core.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	sort.Slice(cenvs, func(i, j int) bool {
		return cenvs[i].Name < cenvs[j].Name
	})
	var r schedulerApi.ProfileTemplate
	r.Container = &schedulerApi.ProfileContainerTemplate{Containers: make(schedulerContainerApi.Containers)}
	r.Container.Containers.ExtendContainers(&schedulerContainerApi.Container{
		Environments: &schedulerContainerResourcesApi.Environments{
			EnvFrom: []core.EnvFromSource{
				{
					SecretRef: &core.SecretEnvSource{
						LocalObjectReference: core.LocalObjectReference{
							Name: secret.GetName(),
						},
						Optional: util.NewType(false),
					},
				},
			},
			Env: cenvs,
		},
	}, containers...)
	return &r, nil
}
