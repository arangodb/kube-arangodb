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
)

// AuthenticationSpec contains the specification to authenticate the destination syncmaster
// with the source syncmasters.
type AuthenticationSpec struct {
	// ClientAuthSecretName holds the name of a Secret containing a client authentication keyfile.
	ClientAuthSecretName *string `json:"clientAuthSecretName,omitempty"`
}

// GetClientAuthSecretName returns the value of clientAuthSecretName.
func (s AuthenticationSpec) GetClientAuthSecretName() string {
	return util.StringOrDefault(s.ClientAuthSecretName)
}

// Validate the given spec, returning an error on validation
// problems or nil if all ok.
func (s AuthenticationSpec) Validate() error {
	if err := k8sutil.ValidateResourceName(s.GetClientAuthSecretName()); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills empty field with default values.
func (s *AuthenticationSpec) SetDefaults() {
}

// SetDefaultsFrom fills empty field with default values from the given source.
func (s *AuthenticationSpec) SetDefaultsFrom(source AuthenticationSpec) {
	if s.ClientAuthSecretName == nil {
		s.ClientAuthSecretName = util.NewStringOrNil(source.ClientAuthSecretName)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s AuthenticationSpec) ResetImmutableFields(target *AuthenticationSpec, fieldPrefix string) []string {
	var result []string
	return result
}
