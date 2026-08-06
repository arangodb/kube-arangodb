//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

// IntegrationLinkV1 enables the Link V1 integration on the sidecar.
// The connector ID is provided at runtime via env var set by the link chart's ArangoProfile.
// The internal address uses the standard INTEGRATION_SERVICE_ADDRESS.
type IntegrationLinkV1 struct {
	Core *Core
}

func NewIntegrationLinkV1() IntegrationLinkV1 {
	t := true
	return IntegrationLinkV1{
		Core: &Core{
			Internal: &t,
			External: &t,
		},
	}
}

func (i IntegrationLinkV1) Name() []string {
	return []string{"LINK", "V1"}
}

func (i IntegrationLinkV1) Validate() error {
	return nil
}

func (i IntegrationLinkV1) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_LINK_V1",
			Value: "true",
		},
		{
			// Enable the external gRPC listener for LinkV1External.
			Name:  "SERVICES_EXTERNAL_ENABLED",
			Value: "true",
		},
		{
			// Enable the external HTTP gateway so the ArangoRoute (HTTP) can reach the link API.
			// Authentication is handled by the gateway (ArangoRoute) before requests
			// reach the external listener. Direct access should be restricted via NetworkPolicy.
			Name:  "SERVICES_EXTERNAL_GATEWAY_ENABLED",
			Value: "true",
		},
	}

	return i.Core.Envs(i, envs...), nil
}

func (i IntegrationLinkV1) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i IntegrationLinkV1) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	return nil, nil, nil
}
