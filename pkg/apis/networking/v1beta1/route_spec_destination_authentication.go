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

package v1beta1

import (
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

type ArangoRouteSpecDestinationAuthentication struct {
	// PassMode define authorization details pass mode when authorization was successful
	// +doc/enum: override|Generates new token for the user
	// +doc/enum: pass|Pass token provided by the user
	// +doc/enum: remove|Removes authorization details from the request
	PassMode *ArangoRouteSpecAuthenticationPassMode `json:"passMode,omitempty"`

	// Type of the authentication
	// +doc/enum: optional|Authentication is header is validated and passed to the service. In case if is unauthorized, requests is still passed
	// +doc/enum: required|Authentication is header is validated and passed to the service. In case if is unauthorized, returns 403
	Type *ArangoRouteSpecAuthenticationType `json:"type,omitempty"`
}

func (a *ArangoRouteSpecDestinationAuthentication) GetType() ArangoRouteSpecAuthenticationType {
	if a == nil {
		return ArangoRouteSpecAuthenticationTypeOptional
	}

	return a.Type.Get()
}

func (a *ArangoRouteSpecDestinationAuthentication) GetPassMode() ArangoRouteSpecAuthenticationPassMode {
	if a == nil {
		return ArangoRouteSpecAuthenticationPassModeOverride
	}

	return a.PassMode.Get()
}

func (a *ArangoRouteSpecDestinationAuthentication) Validate() error {
	if a == nil {
		return nil
	}

	return shared.WithErrors(
		shared.ValidateOptionalInterfacePath("type", a.Type),
		shared.ValidateOptionalInterfacePath("passMode", a.PassMode),
	)
}
