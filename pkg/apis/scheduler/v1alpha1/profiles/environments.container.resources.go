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

package profiles

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	schedulerApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	schedulerContainerApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container"
	schedulerContainerResourcesApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
)

var (
	divisor1m  = resource.MustParse("1m")
	divisor1Mi = resource.MustParse("1Mi")
)

func ContainerResourceEnvironments() *schedulerApiv1alpha1.ProfileTemplate {
	return containerResourceEnvironments.DeepCopy()
}

var containerResourceEnvironments = schedulerApiv1alpha1.ProfileTemplate{
	Container: &schedulerApiv1alpha1.ProfileContainerTemplate{
		All: &schedulerContainerApiv1alpha1.Generic{
			Environments: &schedulerContainerResourcesApiv1alpha1.Environments{
				Env: []core.EnvVar{
					{
						Name: "CONTAINER_CPU_REQUESTS",
						ValueFrom: &core.EnvVarSource{
							ResourceFieldRef: &core.ResourceFieldSelector{
								Resource: "requests.cpu",
								Divisor:  divisor1m,
							},
						},
					},
					{
						Name: "CONTAINER_MEMORY_REQUESTS",
						ValueFrom: &core.EnvVarSource{
							ResourceFieldRef: &core.ResourceFieldSelector{
								Resource: "requests.memory",
								Divisor:  divisor1Mi,
							},
						},
					},
					{
						Name: "CONTAINER_CPU_LIMITS",
						ValueFrom: &core.EnvVarSource{
							ResourceFieldRef: &core.ResourceFieldSelector{
								Resource: "limits.cpu",
								Divisor:  divisor1m,
							},
						},
					},
					{
						Name: "CONTAINER_MEMORY_LIMITS",
						ValueFrom: &core.EnvVarSource{
							ResourceFieldRef: &core.ResourceFieldSelector{
								Resource: "limits.memory",
								Divisor:  divisor1Mi,
							},
						},
					},
				},
			},
		},
	},
}
