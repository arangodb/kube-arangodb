//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type DeploymentSpecGatewayAuthenticationType string

func (d *DeploymentSpecGatewayAuthenticationType) Validate() error {
	if d == nil {
		return nil
	}

	switch v := *d; v {
	case DeploymentSpecGatewayAuthenticationTypeOpenID:
		return nil
	default:
		return errors.Errorf("Invalid AuthenticationType `%s`", v)
	}
}

const (
	DeploymentSpecGatewayAuthenticationTypeOpenID DeploymentSpecGatewayAuthenticationType = "OpenID"
)

type DeploymentSpecGatewayAuthentication struct {
	// Type defines the Authentication Type
	// +doc/enum: OpenID|Configure OpenID Authentication Type
	Type DeploymentSpecGatewayAuthenticationType `json:"type"`

	// Secret defines the secret with the integration configuration
	Secret *sharedApi.Object `json:"secret,omitempty"`
}

func (d *DeploymentSpecGatewayAuthentication) Validate() error {
	if d == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("type", d.Type.Validate()),
		shared.PrefixResourceError("secret", d.Secret.Validate()),
	)
}
