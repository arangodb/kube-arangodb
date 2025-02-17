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

package v1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
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
	Image *string `json:"image,omitempty"`

	// Timeout defines default timeout for the upstream actions (if not overridden)
	// +doc/type: string
	// +doc/default: 1m0s
	Timeout *meta.Duration `json:"timeout,omitempty"`
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

// GetTimeout returns default gateway timeout.
func (d *DeploymentSpecGateway) GetTimeout() meta.Duration {
	if d == nil || d.Timeout == nil {
		return meta.Duration{
			Duration: constants.DefaultEnvoyUpstreamTimeout,
		}
	}

	return *d.Timeout
}

// Validate the given spec
func (d *DeploymentSpecGateway) Validate() error {
	if d == nil {
		d = &DeploymentSpecGateway{}
	}

	return shared.WithErrors(
		shared.PrefixResourceErrorFunc("timeout", func() error {
			if t := d.GetTimeout(); t.Duration < constants.MinEnvoyUpstreamTimeout {
				return errors.Errorf("Timeout lower than %s not allowed", constants.MinEnvoyUpstreamTimeout.String())
			} else if t.Duration > constants.MaxEnvoyUpstreamTimeout {
				return errors.Errorf("Timeout greater than %s not allowed", constants.MaxEnvoyUpstreamTimeout.String())
			}
			return nil
		}),
	)
}

// GetImage returns the image to use for the gateway.
func (d *DeploymentSpecGateway) GetImage() string {
	return util.TypeOrDefault[string](d.Image)
}
