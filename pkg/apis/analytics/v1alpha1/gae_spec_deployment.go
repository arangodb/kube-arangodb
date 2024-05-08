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

package v1alpha1

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
)

const (
	GraphAnalyticsEngineSpecDeploymentApi = "api"

	GraphAnalyticsEngineDeploymentComponentDefaultPort = 8502
)

type GraphAnalyticsEngineSpecDeployment struct {
	// Service defines how components will be exposed
	Service *GraphAnalyticsEngineSpecDeploymentService `json:"service,omitempty"`

	// TLS defined TLS Settings
	TLS *sharedApi.TLS `json:"tls,omitempty"`

	// Port defines on which port the container will be listening for connections
	Port *int32 `json:"port,omitempty"`
}

func (g *GraphAnalyticsEngineSpecDeployment) GetPort(def int32) int32 {
	if g == nil || g.Port == nil {
		return def
	}
	return *g.Port
}

func (g *GraphAnalyticsEngineSpecDeployment) GetService() *GraphAnalyticsEngineSpecDeploymentService {
	if g == nil {
		return nil
	}
	return g.Service
}

func (g *GraphAnalyticsEngineSpecDeployment) GetTLS() *sharedApi.TLS {
	if g == nil {
		return nil
	}
	return g.TLS
}

func (g *GraphAnalyticsEngineSpecDeployment) Validate() error {
	if g == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("service", shared.ValidateOptional(g.GetService(), func(s GraphAnalyticsEngineSpecDeploymentService) error { return s.Validate() })),
	)
}
