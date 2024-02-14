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

package ml

import (
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
)

func AddJWTFolderToPod(deployment *api.ArangoDeployment, spec *core.PodTemplateSpec, integrationContainer *core.Container) {
	if deployment.GetAcceptedSpec().Authentication.IsAuthenticated() {
		spec.Spec.Volumes = append(spec.Spec.Volumes, core.Volume{
			Name: shared.ClusterJWTSecretVolumeName,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: pod.JWTSecretFolder(deployment.GetName()),
				},
			},
		})
	} else {
		spec.Spec.Volumes = append(spec.Spec.Volumes, core.Volume{
			Name: shared.ClusterJWTSecretVolumeName,
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		})
	}

	integrationContainer.VolumeMounts = append(integrationContainer.VolumeMounts, core.VolumeMount{
		Name:      shared.ClusterJWTSecretVolumeName,
		ReadOnly:  true,
		MountPath: shared.ClusterJWTSecretVolumeMountDir,
	})
}
