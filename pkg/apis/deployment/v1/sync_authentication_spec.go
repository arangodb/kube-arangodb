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

package v1

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// SyncAuthenticationSpec holds dc2dc sync authentication specific configuration settings
type SyncAuthenticationSpec struct {
	JWTSecretName      *string `json:"jwtSecretName,omitempty"`      // JWT secret for sync masters
	ClientCASecretName *string `json:"clientCASecretName,omitempty"` // Secret containing client authentication CA
}

// GetJWTSecretName returns the value of jwtSecretName.
func (s SyncAuthenticationSpec) GetJWTSecretName() string {
	return util.TypeOrDefault[string](s.JWTSecretName)
}

// GetClientCASecretName returns the value of clientCASecretName.
func (s SyncAuthenticationSpec) GetClientCASecretName() string {
	return util.TypeOrDefault[string](s.ClientCASecretName)
}

// Validate the given spec
func (s SyncAuthenticationSpec) Validate() error {
	if err := shared.ValidateResourceName(s.GetJWTSecretName()); err != nil {
		return errors.WithStack(err)
	}
	if err := shared.ValidateResourceName(s.GetClientCASecretName()); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *SyncAuthenticationSpec) SetDefaults(defaultJWTSecretName, defaultClientCASecretName string) {
	if s.GetJWTSecretName() == "" {
		// Note that we don't check for nil here, since even a specified, but empty
		// string should result in the default value.
		s.JWTSecretName = util.NewType[string](defaultJWTSecretName)
	}
	if s.GetClientCASecretName() == "" {
		// Note that we don't check for nil here, since even a specified, but empty
		// string should result in the default value.
		s.ClientCASecretName = util.NewType[string](defaultClientCASecretName)
	}
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *SyncAuthenticationSpec) SetDefaultsFrom(source SyncAuthenticationSpec) {
	if s.JWTSecretName == nil {
		s.JWTSecretName = util.NewTypeOrNil[string](source.JWTSecretName)
	}
	if s.ClientCASecretName == nil {
		s.ClientCASecretName = util.NewTypeOrNil[string](source.ClientCASecretName)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to given field prefix.
func (s SyncAuthenticationSpec) ResetImmutableFields(fieldPrefix string, target *SyncAuthenticationSpec) []string {
	var resetFields []string
	return resetFields
}
