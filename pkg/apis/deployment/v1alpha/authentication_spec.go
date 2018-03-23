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
	"github.com/pkg/errors"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// AuthenticationSpec holds authentication specific configuration settings
type AuthenticationSpec struct {
	JWTSecretName *string `json:"jwtSecretName,omitempty"`
}

const (
	// JWTSecretNameDisabled is the value of JWTSecretName to use for disabling authentication.
	JWTSecretNameDisabled = "None"
)

// GetJWTSecretName returns the value of jwtSecretName.
func (s AuthenticationSpec) GetJWTSecretName() string {
	return util.StringOrDefault(s.JWTSecretName)
}

// IsAuthenticated returns true if authentication is enabled.
// Returns false other (when JWTSecretName == "None").
func (s AuthenticationSpec) IsAuthenticated() bool {
	return s.GetJWTSecretName() != JWTSecretNameDisabled
}

// Validate the given spec
func (s AuthenticationSpec) Validate(required bool) error {
	if required && !s.IsAuthenticated() {
		return maskAny(errors.Wrap(ValidationError, "JWT secret is required"))
	}
	if s.IsAuthenticated() {
		if err := k8sutil.ValidateResourceName(s.GetJWTSecretName()); err != nil {
			return maskAny(err)
		}
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *AuthenticationSpec) SetDefaults(defaultJWTSecretName string) {
	if s.GetJWTSecretName() == "" {
		// Note that we don't check for nil here, since even a specified, but empty
		// string should result in the default value.
		s.JWTSecretName = util.NewString(defaultJWTSecretName)
	}
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *AuthenticationSpec) SetDefaultsFrom(source AuthenticationSpec) {
	if s.JWTSecretName == nil {
		s.JWTSecretName = util.NewStringOrNil(source.JWTSecretName)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to given field prefix.
func (s AuthenticationSpec) ResetImmutableFields(fieldPrefix string, target *AuthenticationSpec) []string {
	var resetFields []string
	if s.IsAuthenticated() != target.IsAuthenticated() {
		// Note: You can change the name, but not from empty to non-empty (or reverse).
		target.JWTSecretName = util.NewStringOrNil(s.JWTSecretName)
		resetFields = append(resetFields, fieldPrefix+".jwtSecretName")
	}
	return resetFields
}
