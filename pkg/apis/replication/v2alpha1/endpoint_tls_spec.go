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

// EndpointTLSSpec contains the specification regarding the TLS connection to the syncmasters
// in either source or destination endpoint.
type EndpointTLSSpec struct {
	// CASecretName holds the name of a Secret containing a ca.crt public key for TLS validation.
	CASecretName *string `json:"caSecretName,omitempty"`
}

// GetCASecretName returns the value of caSecretName.
func (s EndpointTLSSpec) GetCASecretName() string {
	return util.TypeOrDefault[string](s.CASecretName)
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s EndpointTLSSpec) Validate(caSecretNameRequired bool) error {
	if err := shared.ValidateOptionalResourceName(s.GetCASecretName()); err != nil {
		return errors.WithStack(err)
	}
	if caSecretNameRequired && s.GetCASecretName() == "" {
		return errors.WithStack(errors.Wrapf(ValidationError, "Provide a caSecretName"))
	}
	return nil
}

// SetDefaults fills empty field with default values.
func (s *EndpointTLSSpec) SetDefaults() {
}

// SetDefaultsFrom fills empty field with default values from the given source.
func (s *EndpointTLSSpec) SetDefaultsFrom(source EndpointTLSSpec) {
	if s.CASecretName == nil {
		s.CASecretName = util.NewTypeOrNil[string](source.CASecretName)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s EndpointTLSSpec) ResetImmutableFields(target *EndpointTLSSpec, fieldPrefix string) []string {
	var result []string
	return result
}
