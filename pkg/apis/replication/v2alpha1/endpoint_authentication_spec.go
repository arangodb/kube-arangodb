//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// EndpointAuthenticationSpec contains the specification to authentication with the syncmasters
// in either source or destination endpoint.
type EndpointAuthenticationSpec struct {
	// KeyfileSecretName holds the name of a Secret containing a client authentication
	// certificate formatted at keyfile in a `tls.keyfile` field.
	KeyfileSecretName *string `json:"keyfileSecretName,omitempty"`
	// UserSecretName holds the name of a Secret containing a `username` & `password`
	// field used for basic authentication.
	// The user identified by the username must have write access in the `_system` database
	// of the ArangoDB cluster at the endpoint.
	UserSecretName *string `json:"userSecretName,omitempty"`
}

// GetKeyfileSecretName returns the value of keyfileSecretName.
func (s EndpointAuthenticationSpec) GetKeyfileSecretName() string {
	return util.TypeOrDefault[string](s.KeyfileSecretName)
}

// GetUserSecretName returns the value of userSecretName.
func (s EndpointAuthenticationSpec) GetUserSecretName() string {
	return util.TypeOrDefault[string](s.UserSecretName)
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s EndpointAuthenticationSpec) Validate(keyfileSecretNameRequired bool) error {
	if err := shared.ValidateOptionalResourceName(s.GetKeyfileSecretName()); err != nil {
		return errors.WithStack(err)
	}
	if err := shared.ValidateOptionalResourceName(s.GetUserSecretName()); err != nil {
		return errors.WithStack(err)
	}
	if keyfileSecretNameRequired && s.GetKeyfileSecretName() == "" {
		return errors.WithStack(errors.Wrapf(ValidationError, "Provide a keyfileSecretName"))
	}
	return nil
}

// SetDefaults fills empty field with default values.
func (s *EndpointAuthenticationSpec) SetDefaults() {
}

// SetDefaultsFrom fills empty field with default values from the given source.
func (s *EndpointAuthenticationSpec) SetDefaultsFrom(source EndpointAuthenticationSpec) {
	if s.KeyfileSecretName == nil {
		s.KeyfileSecretName = util.NewTypeOrNil[string](source.KeyfileSecretName)
	}
	if s.UserSecretName == nil {
		s.UserSecretName = util.NewTypeOrNil[string](source.UserSecretName)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s EndpointAuthenticationSpec) ResetImmutableFields(target *EndpointAuthenticationSpec, fieldPrefix string) []string {
	var result []string
	return result
}
