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

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod"
	schedulerPodResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func baseAsTemplate(in *pbSchedulerV1.Spec) *schedulerApi.ProfileTemplate {
	containers := schedulerContainerApi.Containers{}

	for n, c := range in.Containers {
		if c == nil {
			continue
		}

		var container schedulerContainerApi.Container

		if image := c.Image; image != nil {
			container.Image = &schedulerContainerResourcesApi.Image{
				Image: c.Image,
			}
		}

		if len(c.Args) > 0 {
			container.Core = &schedulerContainerResourcesApi.Core{
				Args: c.Args,
			}
		}

		if len(c.EnvironmentVariables) > 0 {
			container.Environments = &schedulerContainerResourcesApi.Environments{}

			for k, v := range c.EnvironmentVariables {
				container.Env = append(container.Env, core.EnvVar{
					Name:  k,
					Value: v,
				})
			}
		}

		containers[n] = container
	}

	var t = schedulerApi.ProfileTemplate{
		Priority: util.NewType(math.MaxInt),
		Pod:      &schedulerPodApi.Pod{},
		Container: &schedulerApi.ProfileContainerTemplate{
			Containers: containers,
		},
	}

	if job := in.Job; job != nil {
		t.Pod.Metadata = &schedulerPodResourcesApi.Metadata{
			Labels: job.Labels,
		}
	}

	return &t
}
