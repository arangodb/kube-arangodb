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
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

func GetJWTAuthFileTokenPath(prefix string) string {
	base := "/etc/arangodb/jwt"
	if prefix == "" {
		return base
	}

	return fmt.Sprintf("%s-%s", base, prefix)
}

func AddJWTAuthFileToContainers(ext *mlApi.ArangoMLExtension, deployment *api.ArangoDeployment, spec *core.PodTemplateSpec, containers ...*core.Container) {
	authSpec := deployment.GetAcceptedSpec().Authentication
	if !authSpec.IsAuthenticated() {
		return
	}

	if ext.GetStatus().ArangoDB == nil {
		// not ready yet, skip for now
		return
	}

	mountJWTTokenSecret("", ext.GetStatus().ArangoDB.JWTTokenSecret, spec, containers...)
	mountJWTTokenSecret("METADATA", ext.GetStatus().MetadataService.JWTTokenSecret, spec, containers...)
}

// mountJWTTokenSecret is assuming that prefix contains only alphanumeric symbols and/or '-'
func mountJWTTokenSecret(prefix string, secret *sharedApi.Object, spec *core.PodTemplateSpec, containers ...*core.Container) {
	if secret.IsEmpty() {
		return
	}

	mountName := "deployment-auth-jwt"
	if prefix != "" {
		mountName = fmt.Sprintf("%s-%s", mountName, strings.ToLower(prefix))
	}
	spec.Spec.Volumes = append(spec.Spec.Volumes, core.Volume{
		Name: mountName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: secret.GetName(),
			},
		},
	})

	for _, container := range containers {
		container.VolumeMounts = append(container.VolumeMounts, core.VolumeMount{
			Name:      mountName,
			ReadOnly:  true,
			MountPath: GetJWTAuthFileTokenPath(prefix),
		})
	}
}
