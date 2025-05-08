//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"fmt"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type IntegrationEnvoyV3 struct {
	Core           *Core
	DeploymentName string
	Spec           api.DeploymentSpec
}

func (i IntegrationEnvoyV3) Name() []string {
	return []string{"ENVOY", "AUTH", "V3"}
}

func (i IntegrationEnvoyV3) Validate() error {
	return nil
}

func (i IntegrationEnvoyV3) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_ENVOY_AUTH_V3",
			Value: "true",
		},
		{
			Name:  "INTEGRATION_ENVOY_AUTH_V3_EXTENSIONS_COOKIE_JWT",
			Value: util.BoolSwitch(i.Spec.Gateway.IsCookiesSupportEnabled(), "true", "false"),
		},
		{
			Name:  "INTEGRATION_ENVOY_AUTH_V3_EXTENSIONS_USERS_CREATE",
			Value: util.BoolSwitch(i.Spec.Gateway.IsCreateUsersEnabled(), "true", "false"),
		},
	}

	if gw := i.Spec.Gateway; gw != nil {
		if auth := gw.Authentication; auth != nil {
			if obj := auth.Secret; obj != nil {
				envs = append(envs,
					core.EnvVar{
						Name:  "INTEGRATION_ENVOY_AUTH_V3_AUTH_ENABLED",
						Value: "true",
					},
					core.EnvVar{
						Name:  "INTEGRATION_ENVOY_AUTH_V3_AUTH_TYPE",
						Value: string(auth.Type),
					},
					core.EnvVar{
						Name:  "INTEGRATION_ENVOY_AUTH_V3_AUTH_PATH",
						Value: fmt.Sprintf("%sconfig", constants.MemberAuthVolumeMountDir),
					},
				)
			}
		}
	}

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationEnvoyV3) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i IntegrationEnvoyV3) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	volumes := pod.NewVolumes()

	if gw := i.Spec.Gateway; gw != nil {
		if auth := gw.Authentication; auth != nil {
			if obj := auth.Secret; obj != nil {
				volumes.AddVolume(k8sutil.CreateVolumeWithSecret(constants.MemberAuthVolumeName, obj.GetName()))

				volumes.AddVolumeMount(core.VolumeMount{
					Name:      constants.MemberAuthVolumeName,
					MountPath: constants.MemberAuthVolumeMountDir,
					ReadOnly:  true,
				})
			}
		}
	}

	return volumes.Volumes(), volumes.VolumeMounts(), nil
}
