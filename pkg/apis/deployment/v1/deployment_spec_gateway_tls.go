//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// DeploymentSpecGatewayTLS configures the gateway's client-facing (downstream) TLS listener,
// independently of the deployment (arangod) TLS. When enabled (and the gateway-tls operator feature is
// on), the gateway serves TLS to external clients even if `spec.tls` is disabled. The gateway->arangod
// upstream connection is unaffected and still follows `spec.tls`.
type DeploymentSpecGatewayTLS struct {
	// Enabled turns on the gateway's client-facing TLS listener. Takes effect only when the operator
	// `gateway-tls` feature is enabled. When `spec.tls` is already secure the gateway serves TLS
	// regardless of this setting.
	// +doc/default: false
	Enabled *bool `json:"enabled,omitempty"`

	// CASecretName optionally names a user-provided CA secret (containing `ca.crt` and `ca.key`) used to
	// sign the gateway serving certificate. When unset, the operator generates and manages a self-signed
	// CA for the gateway.
	CASecretName *string `json:"caSecretName,omitempty"`

	// AltNames holds additional subject alternative names (DNS names / IPs) added to the generated
	// gateway serving certificate.
	AltNames []string `json:"altNames,omitempty"`
}

// IsEnabled returns whether the gateway client-facing TLS listener is requested.
func (d *DeploymentSpecGatewayTLS) IsEnabled() bool {
	if d == nil || d.Enabled == nil {
		return false
	}

	return *d.Enabled
}

// GetCASecretName returns the user-provided CA secret name, or an empty string when the operator should
// generate and manage the CA itself.
func (d *DeploymentSpecGatewayTLS) GetCASecretName() string {
	if d == nil || d.CASecretName == nil {
		return ""
	}

	return *d.CASecretName
}

// HasCustomCA returns whether a user-provided CA secret is configured.
func (d *DeploymentSpecGatewayTLS) HasCustomCA() bool {
	return d.GetCASecretName() != ""
}

// GetAltNames returns the additional alt names for the generated gateway certificate.
func (d *DeploymentSpecGatewayTLS) GetAltNames() []string {
	if d == nil {
		return nil
	}

	return d.AltNames
}

// AsTLSSpec builds the TLSSpec used by the operator to generate and serve the gateway certificate. When
// no custom CA is configured, defaultCASecretName (the operator-managed gateway CA) is used.
func (d *DeploymentSpecGatewayTLS) AsTLSSpec(defaultCASecretName string) TLSSpec {
	ca := d.GetCASecretName()
	if ca == "" {
		ca = defaultCASecretName
	}

	return TLSSpec{
		CASecretName: util.NewType[string](ca),
		AltNames:     d.GetAltNames(),
	}
}

// Validate the given spec.
func (d *DeploymentSpecGatewayTLS) Validate() error {
	if d == nil {
		return nil
	}

	if n := d.GetCASecretName(); n != "" {
		if err := shared.ValidateResourceName(n); err != nil {
			return shared.PrefixResourceError("caSecretName", err)
		}
	}

	return nil
}
