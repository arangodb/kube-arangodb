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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
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
		{
			Name:  "INTEGRATION_ENVOY_AUTH_V3_DATABASE_ENDPOINT",
			Value: k8sutil.ExtendDeploymentClusterDomain(fmt.Sprintf("%s-%s", i.DeploymentName, i.Spec.GetMode().ServingGroup().AsRole()), i.Spec.ClusterDomain),
		},
		{
			Name:  "INTEGRATION_ENVOY_AUTH_V3_DATABASE_PROTO",
			Value: util.BoolSwitch(i.Spec.IsSecure(), "https", "http"),
		},
		{
			Name:  "INTEGRATION_ENVOY_AUTH_V3_DATABASE_PORT",
			Value: fmt.Sprintf("%d", shared.ArangoPort),
		},
	}

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationEnvoyV3) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i IntegrationEnvoyV3) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	return nil, nil, nil
}
