//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package v1alpha

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
)

// EndpointAuthenticationSpec contains the specification to authentication with the syncmasters
// in either source or destination endpoint.
type EndpointAuthenticationSpec struct {
	// JWTSecretName holds the name of a Secret containing a JWT token.
	JWTSecretName *string `json:"jwtSecretName,omitempty"`
}

// GetJWTSecretName returns the value of jwtSecretName.
func (s EndpointAuthenticationSpec) GetJWTSecretName() string {
	return util.StringOrDefault(s.JWTSecretName)
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s EndpointAuthenticationSpec) Validate(jwtSecretNameRequired bool) error {
	if err := k8sutil.ValidateOptionalResourceName(s.GetJWTSecretName()); err != nil {
		return maskAny(err)
	}
	if jwtSecretNameRequired && s.GetJWTSecretName() == "" {
		return maskAny(errors.Wrapf(ValidationError, "Provide a jwtSecretName"))
	}
	return nil
}

// SetDefaults fills empty field with default values.
func (s *EndpointAuthenticationSpec) SetDefaults() {
}

// SetDefaultsFrom fills empty field with default values from the given source.
func (s *EndpointAuthenticationSpec) SetDefaultsFrom(source EndpointAuthenticationSpec) {
	if s.JWTSecretName == nil {
		s.JWTSecretName = util.NewStringOrNil(source.JWTSecretName)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s EndpointAuthenticationSpec) ResetImmutableFields(target *EndpointAuthenticationSpec, fieldPrefix string) []string {
	var result []string
	return result
}
