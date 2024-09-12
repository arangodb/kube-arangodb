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
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type DeploymentSpecGateway struct {
	// Enabled setting enables/disables support for gateway in the cluster.
	// When enabled, the cluster will contain a number of `gateway` servers.
	// +doc/default: false
	Enabled *bool `json:"enabled,omitempty"`

	// Dynamic setting enables/disables support dynamic configuration of the gateway in the cluster.
	// When enabled, gateway config will be reloaded by ConfigMap live updates.
	// +doc/default: false
	Dynamic *bool `json:"dynamic,omitempty"`

	// Image is the image to use for the gateway.
	// By default, the image is determined by the operator.
	Image *string `json:"image"`
}

// IsEnabled returns whether the gateway is enabled.
func (d *DeploymentSpecGateway) IsEnabled() bool {
	if d == nil || d.Enabled == nil {
		return false
	}

	return *d.Enabled
}

// IsDynamic returns whether the gateway dynamic config is enabled.
func (d *DeploymentSpecGateway) IsDynamic() bool {
	if d == nil || d.Dynamic == nil {
		return false
	}

	return *d.Dynamic
}

// Validate the given spec
func (d *DeploymentSpecGateway) Validate() error {
	return nil
}

// GetImage returns the image to use for the gateway.
func (d *DeploymentSpecGateway) GetImage() string {
	return util.TypeOrDefault[string](d.Image)
}
