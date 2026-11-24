//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package sidecar

import (
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

type IntegrationAuthenticationV1 struct {
	Core *Core

	DeploymentName string
	Spec           api.DeploymentSpec
}

func (i IntegrationAuthenticationV1) Name() []string {
	return []string{"AUTHENTICATION", "V1"}
}

func (i IntegrationAuthenticationV1) Validate() error {
	return nil
}

func (i IntegrationAuthenticationV1) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_AUTHENTICATION_V1",
			Value: "true",
		},
		{
			Name:  "INTEGRATION_AUTHENTICATION_V1_ENABLED",
			Value: util.BoolSwitch(i.Spec.IsAuthenticated(), "true", "false"),
		},
		{
			Name:  "INTEGRATION_AUTHENTICATION_V1_PATH",
			Value: shared.ClusterJWTSecretVolumeMountDir,
		},
	}

	if i.Spec.IsAuthenticated() {
		// Point the integration's client-connection authenticator at the cluster JWT secret this
		// integration mounts (see Volumes), so it can sign a local superuser token for the central
		// authorization service. It is set here, gated on the same condition as the mount and scoped to
		// this integration's profile, rather than on the shared integration container: otherwise the env
		// would still be present when this local integration is dropped in favour of the central service
		// (and the folder is not mounted), making every proxied call fail with Unauthenticated. With no
		// token source the client simply forwards without authentication.
		envs = append(envs, core.EnvVar{
			Name:  utilConstants.INTEGRATION_ARANGO_JWT_FOLDER.String(),
			Value: shared.ClusterJWTSecretVolumeMountDir,
		})
	}

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationAuthenticationV1) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i IntegrationAuthenticationV1) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	if i.Spec.IsAuthenticated() {
		return []core.Volume{
			{
				Name: shared.ClusterJWTSecretVolumeName,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: pod.JWTSecretFolder(i.DeploymentName),
					},
				},
			},
		}, []core.VolumeMount{
			{
				Name:      shared.ClusterJWTSecretVolumeName,
				ReadOnly:  true,
				MountPath: shared.ClusterJWTSecretVolumeMountDir,
			},
		}, nil
	}

	return nil, nil, nil
}
