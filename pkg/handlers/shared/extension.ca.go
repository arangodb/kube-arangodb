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
	"fmt"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod"
	schedulerPodResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod/resources"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

func AddDeploymentCATemplate(deployment *api.ArangoDeployment, containers ...string) *schedulerApi.ProfileTemplate {
	if !deployment.GetAcceptedSpec().TLS.IsSecure() {
		return nil
	}
	pt := &schedulerApi.ProfileTemplate{
		Pod: &schedulerPodApi.Pod{
			Volumes: &schedulerPodResourcesApi.Volumes{
				Volumes: []core.Volume{
					{
						Name: "deployment-ca",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: resources.GetCASecretName(deployment),
							},
						},
					},
				},
			},
		},
		Container: &schedulerApi.ProfileContainerTemplate{
			Containers: make(schedulerContainerApi.Containers),
		},
	}
	pt.Container.Containers.ExtendContainers(&schedulerContainerApi.Container{
		VolumeMounts: &schedulerContainerResourcesApi.VolumeMounts{
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "deployment-ca",
					ReadOnly:  true,
					MountPath: "/etc/arangodb/tls",
				},
			},
		},
	}, containers...)
	return pt
}
func AddTLSSecretTemplate(tls *sharedApi.Object, containers ...string) *schedulerApi.ProfileTemplate {
	pt := &schedulerApi.ProfileTemplate{
		Pod: &schedulerPodApi.Pod{
			Volumes: &schedulerPodResourcesApi.Volumes{
				Volumes: []core.Volume{
					{
						Name: "keyfile",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: tls.GetName(),
							},
						},
					},
				},
			},
		},
		Container: &schedulerApi.ProfileContainerTemplate{
			Containers: make(schedulerContainerApi.Containers),
		},
	}
	pt.Container.Containers.ExtendContainers(&schedulerContainerApi.Container{
		VolumeMounts: &schedulerContainerResourcesApi.VolumeMounts{
			VolumeMounts: []core.VolumeMount{
				{
					Name:      "keyfile",
					ReadOnly:  true,
					MountPath: "/etc/arangodb/keyfile",
				},
			},
		},
		Environments: &schedulerContainerResourcesApi.Environments{
			Env: []core.EnvVar{
				{
					Name:  "SECRET_CERTFILE_CHECKSUM",
					Value: tls.GetChecksum(),
				},
				{
					Name:  "SERVER_CERTFILE",
					Value: fmt.Sprintf("/etc/arangodb/keyfile/%s", utilConstants.SecretTLSKeyfile),
				},
			},
		},
	}, containers...)
	return pt
}
