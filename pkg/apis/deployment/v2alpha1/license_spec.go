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
)

// LicenseSpec holds the license related information
type LicenseSpec struct {
	SecretName *string `json:"secretName,omitempty"`
}

// HasSecretName returns true if a license key secret name was set
func (s LicenseSpec) HasSecretName() bool {
	return s.SecretName != nil
}

// GetSecretName returns the license key if set. Empty string otherwise.
func (s LicenseSpec) GetSecretName() string {
	return util.TypeOrDefault[string](s.SecretName)
}

// Validate validates the LicenseSpec
func (s LicenseSpec) Validate() error {
	if s.HasSecretName() {
		if err := shared.ValidateResourceName(s.GetSecretName()); err != nil {
			return err
		}
	}

	return nil
}

// SetDefaultsFrom fills all values not set in s with values from other
func (s *LicenseSpec) SetDefaultsFrom(other LicenseSpec) {
	if !s.HasSecretName() {
		s.SecretName = util.NewTypeOrNil[string](other.SecretName)
	}
}
