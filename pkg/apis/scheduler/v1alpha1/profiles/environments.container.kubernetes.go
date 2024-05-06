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

	schedulerApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	schedulerContainerApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container"
	schedulerContainerResourcesApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
)

func ContainerKubernetesEnvironments() *schedulerApiv1alpha1.ProfileTemplate {
	return containerKubernetesEnvironments.DeepCopy()
}

var containerKubernetesEnvironments = schedulerApiv1alpha1.ProfileTemplate{
	Container: &schedulerApiv1alpha1.ProfileContainerTemplate{
		All: &schedulerContainerApiv1alpha1.Generic{
			Environments: &schedulerContainerResourcesApiv1alpha1.Environments{
				Env: []core.EnvVar{
					{
						Name: "KUBE_NAMESPACE",
						ValueFrom: &core.EnvVarSource{
							FieldRef: &core.ObjectFieldSelector{
								FieldPath: "metadata.namespace",
							},
						},
					},
					{
						Name: "KUBE_NAME",
						ValueFrom: &core.EnvVarSource{
							FieldRef: &core.ObjectFieldSelector{
								FieldPath: "metadata.name",
							},
						},
					},
					{
						Name: "KUBE_IP",
						ValueFrom: &core.EnvVarSource{
							FieldRef: &core.ObjectFieldSelector{
								FieldPath: "status.podIP",
							},
						},
					},
					{
						Name: "KUBE_SERVICE_ACCOUNT",
						ValueFrom: &core.EnvVarSource{
							FieldRef: &core.ObjectFieldSelector{
								FieldPath: "spec.serviceAccountName",
							},
						},
					},
				},
			},
		},
	},
}
