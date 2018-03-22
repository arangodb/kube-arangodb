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
	"k8s.io/api/core/v1"
)

// SyncSpec holds dc2dc replication specific configuration settings
type SyncSpec struct {
	Enabled         bool          `json:"enabled,omitempty"`
	Image           string        `json:"image,omitempty"`
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`

	Authentication AuthenticationSpec `json:"auth"`
	TLS            TLSSpec            `json:"tls"`
	Monitoring     MonitoringSpec     `json:"monitoring"`
}

// Validate the given spec
func (s SyncSpec) Validate(mode DeploymentMode) error {
	if s.Enabled && !mode.SupportsSync() {
		return maskAny(errors.Wrapf(ValidationError, "Cannot enable sync with mode: '%s'", mode))
	}
	if s.Image == "" {
		return maskAny(errors.Wrapf(ValidationError, "image must be set"))
	}
	if err := s.Authentication.Validate(s.Enabled); err != nil {
		return maskAny(err)
	}
	if s.Enabled {
		if err := s.TLS.Validate(); err != nil {
			return maskAny(err)
		}
	}
	if err := s.Monitoring.Validate(); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *SyncSpec) SetDefaults(defaultImage string, defaulPullPolicy v1.PullPolicy, defaultJWTSecretName, defaultCASecretName string) {
	if s.Image == "" {
		s.Image = defaultImage
	}
	if s.ImagePullPolicy == "" {
		s.ImagePullPolicy = defaulPullPolicy
	}
	s.Authentication.SetDefaults(defaultJWTSecretName)
	s.TLS.SetDefaults(defaultCASecretName)
	s.Monitoring.SetDefaults()
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *SyncSpec) SetDefaultsFrom(source SyncSpec) {
	if s.Image == "" {
		s.Image = source.Image
	}
	if s.ImagePullPolicy == "" {
		s.ImagePullPolicy = source.ImagePullPolicy
	}
	s.Authentication.SetDefaultsFrom(source.Authentication)
	s.TLS.SetDefaultsFrom(source.TLS)
	s.Monitoring.SetDefaultsFrom(source.Monitoring)
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to given field prefix.
func (s SyncSpec) ResetImmutableFields(fieldPrefix string, target *SyncSpec) []string {
	var resetFields []string
	if list := s.Authentication.ResetImmutableFields(fieldPrefix+".auth", &target.Authentication); len(list) > 0 {
		resetFields = append(resetFields, list...)
	}
	return resetFields
}
