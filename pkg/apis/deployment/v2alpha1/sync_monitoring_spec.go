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

// MonitoringSpec holds monitoring specific configuration settings
type MonitoringSpec struct {
	TokenSecretName *string `json:"tokenSecretName,omitempty"`
}

// GetTokenSecretName returns the value of tokenSecretName.
func (s MonitoringSpec) GetTokenSecretName() string {
	return util.TypeOrDefault[string](s.TokenSecretName)
}

// Validate the given spec
func (s MonitoringSpec) Validate() error {
	if err := shared.ValidateOptionalResourceName(s.GetTokenSecretName()); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *MonitoringSpec) SetDefaults(defaultTokenSecretName string) {
	if s.GetTokenSecretName() == "" {
		// Note that we don't check for nil here, since even a specified, but empty
		// string should result in the default value.
		s.TokenSecretName = util.NewType[string](defaultTokenSecretName)
	}
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *MonitoringSpec) SetDefaultsFrom(source MonitoringSpec) {
	if s.TokenSecretName == nil {
		s.TokenSecretName = util.NewTypeOrNil[string](source.TokenSecretName)
	}
}
