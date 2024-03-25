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

package scheduler

import (
	"math"

	core "k8s.io/api/core/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/pod"
	schedulerPodResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/pod/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

const DefaultContainerName = "job"

type Request struct {
	Labels map[string]string

	Profiles []string

	Envs map[string]string

	Container *string

	Image *string

	Args []string
}

func (r Request) AsTemplate() *schedulerApi.ProfileTemplate {
	var container schedulerContainerApi.Container

	if len(r.Envs) > 0 {
		container.Environments = &schedulerContainerResourcesApi.Environments{}

		for k, v := range r.Envs {
			container.Environments.Env = append(container.Environments.Env, core.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	if len(r.Args) > 0 {
		container.Core = &schedulerContainerResourcesApi.Core{
			Args: r.Args,
		}
	}

	if r.Image != nil {
		container.Image = &schedulerContainerResourcesApi.Image{
			Image: util.NewType(util.TypeOrDefault(r.Image)),
		}
	}

	return &schedulerApi.ProfileTemplate{
		Priority: util.NewType(math.MaxInt),
		Pod: &schedulerPodApi.Pod{
			Metadata: &schedulerPodResourcesApi.Metadata{
				Labels: util.MergeMaps(true, r.Labels),
			},
		},
		Container: &schedulerApi.ProfileContainerTemplate{
			Containers: map[string]schedulerContainerApi.Container{
				util.TypeOrDefault(r.Container, DefaultContainerName): container,
			},
		},
	}
}
