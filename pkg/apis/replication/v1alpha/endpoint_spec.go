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
	"net/url"

	"github.com/pkg/errors"
)

// EndpointSpec contains the specification used to reach the syncmasters
// in either source or destination mode.
type EndpointSpec struct {
	MasterEndpoint []string                   `json:"masterEndpoint,omitempty"`
	Authentication EndpointAuthenticationSpec `json:"auth"`
	TLS            EndpointTLSSpec            `json:"tls"`
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s EndpointSpec) Validate() error {
	for _, ep := range s.MasterEndpoint {
		if _, err := url.Parse(ep); err != nil {
			return maskAny(errors.Wrapf(ValidationError, "Invalid master endpoint '%s': %s", ep, err))
		}
	}
	if err := s.Authentication.Validate(); err != nil {
		return maskAny(err)
	}
	if err := s.TLS.Validate(); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills empty field with default values.
func (s *EndpointSpec) SetDefaults() {
	s.Authentication.SetDefaults()
	s.TLS.SetDefaults()
}

// SetDefaultsFrom fills empty field with default values from the given source.
func (s *EndpointSpec) SetDefaultsFrom(source EndpointSpec) {
	s.Authentication.SetDefaultsFrom(source.Authentication)
	s.TLS.SetDefaultsFrom(source.TLS)
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s EndpointSpec) ResetImmutableFields(target *EndpointSpec, fieldPrefix string) []string {
	var result []string
	if list := s.Authentication.ResetImmutableFields(&target.Authentication, fieldPrefix+"auth."); len(list) > 0 {
		result = append(result, list...)
	}
	if list := s.TLS.ResetImmutableFields(&target.TLS, fieldPrefix+"tls."); len(list) > 0 {
		result = append(result, list...)
	}
	return result
}
