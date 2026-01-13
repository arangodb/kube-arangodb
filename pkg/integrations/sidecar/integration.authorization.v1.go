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
)

type IntegrationAuthorizationV1 struct {
	Core *Core
}

func (i IntegrationAuthorizationV1) Name() []string {
	return []string{"AUTHORIZATION", "V1"}
}

func (i IntegrationAuthorizationV1) Validate() error {
	return nil
}

func (i IntegrationAuthorizationV1) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_AUTHENTICATION_V1",
			Value: "true",
		},
	}

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationAuthorizationV1) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i IntegrationAuthorizationV1) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	return nil, nil, nil
}
