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
//

package v1alpha

import "github.com/arangodb/kube-arangodb/pkg/util"

// MetricsSpec contains spec for arangodb exporter
type MetricsSpec struct {
	Enabled *bool   `json:"enabled,omitempty"`
	Image   *string `json:"image,omitempty"`
	//Authentication struct {
	//	// JWTSecretName contains the name of the JWT kubernetes secret used for authentication
	//	JWTSecretName *string `json:"JWTSecretName,omitempty"`
	//} `json:"authentication,omitempty"`
}

// IsEnabled returns whether metrics are enabled or not
func (s *MetricsSpec) IsEnabled() bool {
	return util.BoolOrDefault(s.Enabled, false)
}

// HasImage returns whether a image was specified or not
func (s *MetricsSpec) HasImage() bool {
	return s.Image != nil
}

// GetImage returns the Image or empty string
func (s *MetricsSpec) GetImage() string {
	return util.StringOrDefault(s.Image)
}

// SetDefaults sets default values
func (s *MetricsSpec) SetDefaults() {
	s.Enabled = util.NewBool(false)
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *MetricsSpec) SetDefaultsFrom(source MetricsSpec) {
	if s.Enabled == nil {
		s.Enabled = util.NewBoolOrNil(source.Enabled)
	}
	if s.Image == nil {
		s.Image = util.NewStringOrNil(source.Image)
	}
}

// Validate the given spec
func (s *MetricsSpec) Validate() error {
	return nil
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
func (s SyncSpec) ResetImmutableFields(fieldPrefix string, target *SyncSpec) []string {
	return nil
}
