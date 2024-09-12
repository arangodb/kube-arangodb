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

package v2alpha1

import (
	schedulerIntegrationApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/integration"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type DeploymentSpecIntegration struct {
	// Sidecar define the integration sidecar spec
	Sidecar *schedulerIntegrationApi.Sidecar `json:"sidecar,omitempty"`
}

func (d *DeploymentSpecIntegration) GetSidecar() *schedulerIntegrationApi.Sidecar {
	if d == nil || d.Sidecar == nil {
		return nil
	}
	return d.Sidecar
}

// Validate the given spec
func (d *DeploymentSpecIntegration) Validate() error {
	if d == nil {
		d = &DeploymentSpecIntegration{}
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("sidecar", d.GetSidecar().Validate()),
	)
}
