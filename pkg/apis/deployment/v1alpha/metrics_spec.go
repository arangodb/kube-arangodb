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

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
)

// MetricsAuthenticationSpec contains spec for authentication with arangodb
type MetricsAuthenticationSpec struct {
	// JWTTokenSecretName contains the name of the JWT kubernetes secret used for authentication
	JWTTokenSecretName *string `json:"jwtTokenSecretName,omitempty"`
}

// MetricsMonitorResourceMode specifies the type of ServiceMonitor resource to be installed
type MetricsMonitorResourceMode string

const (
	// MetricsMonitorResourceModeAuto detects if the CRD for ServiceMonitor is available and creates the resource if so (default)
	MetricsMonitorResourceModeAuto MetricsMonitorResourceMode = "Auto"
	// MetricsMonitorResourceModeNone never creates the resource
	MetricsMonitorResourceModeNone MetricsMonitorResourceMode = "None"
	// MetricsMonitorResourceModeAlways always creates the resource and fails if the CRD is not available
	MetricsMonitorResourceModeAlways MetricsMonitorResourceMode = "Always"
)

// IsAuto returns true if mode is MetricsMonitorResourceModeAuto
func (t MetricsMonitorResourceMode) IsAuto() bool {
	return t == MetricsMonitorResourceModeAuto
}

// IsNone returns true if mode is MetricsMonitorResourceModeNone
func (t MetricsMonitorResourceMode) IsNone() bool {
	return t == MetricsMonitorResourceModeNone
}

// IsAlways returns true if mode is MetricsMonitorResourceModeAlways
func (t MetricsMonitorResourceMode) IsAlways() bool {
	return t == MetricsMonitorResourceModeAlways
}

// MetricsPrometheusSpec contains metrics for prometheus
type MetricsPrometheusSpec struct {
	MonitorResource *MetricsMonitorResourceMode `json:"monitorResource,omitempty"`
}

// MetricsSpec contains spec for arangodb exporter
type MetricsSpec struct {
	Enabled        *bool                     `json:"enabled,omitempty"`
	Image          *string                   `json:"image,omitempty"`
	Authentication MetricsAuthenticationSpec `json:"authentication,omitempty"`
	Prometheus     MetricsPrometheusSpec     `json:"prometheus,omitempty"`
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
func (s *MetricsSpec) SetDefaults(defaultTokenName string, isAuthenticated bool) {
	if s.Enabled == nil {
		s.Enabled = util.NewBool(false)
	}
	if s.GetJWTTokenSecretName() == "" {
		s.Authentication.JWTTokenSecretName = util.NewString(defaultTokenName)
	}
	s.Prometheus.SetDefaults()
}

// GetJWTTokenSecretName returns the token secret name or empty string
func (s *MetricsSpec) GetJWTTokenSecretName() string {
	return util.StringOrDefault(s.Authentication.JWTTokenSecretName)
}

// HasJWTTokenSecretName returns true if a secret name was specified
func (s *MetricsSpec) HasJWTTokenSecretName() bool {
	return s.Authentication.JWTTokenSecretName != nil
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *MetricsSpec) SetDefaultsFrom(source MetricsSpec) {
	if s.Enabled == nil {
		s.Enabled = util.NewBoolOrNil(source.Enabled)
	}
	if s.Image == nil {
		s.Image = util.NewStringOrNil(source.Image)
	}
	if s.Authentication.JWTTokenSecretName == nil {
		s.Authentication.JWTTokenSecretName = util.NewStringOrNil(source.Authentication.JWTTokenSecretName)
	}
	s.Prometheus.SetDefaultsFrom(source.Prometheus)
}

// Validate the given spec
func (s *MetricsSpec) Validate() error {

	if s.HasJWTTokenSecretName() {
		if err := k8sutil.ValidateResourceName(s.GetJWTTokenSecretName()); err != nil {
			return err
		}
	}

	if err := s.Prometheus.Validate(); err != nil {
		return err
	}

	return nil
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
func (s MetricsSpec) ResetImmutableFields(fieldPrefix string, target *MetricsSpec) []string {
	return nil
}

// SetDefaults sets default values
func (p *MetricsPrometheusSpec) SetDefaults() {
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (p *MetricsPrometheusSpec) SetDefaultsFrom(source MetricsPrometheusSpec) {
	if p.MonitorResource == nil {
		p.MonitorResource = NewMetricsMonitorResourceModeOrNil(source.MonitorResource)
	}
}

// Validate the given spec
func (p *MetricsPrometheusSpec) Validate() error {
	return p.MonitorResource.Validate()
}

// Validate the given spec
func (t MetricsMonitorResourceMode) Validate() error {
	switch t {
	case MetricsMonitorResourceModeAuto, MetricsMonitorResourceModeNone, MetricsMonitorResourceModeAlways:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown monitor resource mode: '%s'", string(t)))
	}
}

// NewMetricsMonitorResourceMode returns a reference to a string with given value.
func NewMetricsMonitorResourceMode(input MetricsMonitorResourceMode) *MetricsMonitorResourceMode {
	return &input
}

// NewMetricsMonitorResourceModeOrNil returns nil if input is nil, otherwise returns a clone of the given value.
func NewMetricsMonitorResourceModeOrNil(input *MetricsMonitorResourceMode) *MetricsMonitorResourceMode {
	if input == nil {
		return nil
	}

	return NewMetricsMonitorResourceMode(*input)
}

// MetricsMonitorResourceModeOrDefault returns the default value (or empty string) if input is nil, otherwise returns the referenced value.
func MetricsMonitorResourceModeOrDefault(input *MetricsMonitorResourceMode, defaultValue ...MetricsMonitorResourceMode) MetricsMonitorResourceMode {
	if input == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return *input
}

// GetMode returns the given mode or the default value if it is nil
func (t *MetricsMonitorResourceMode) GetMode() MetricsMonitorResourceMode {
	return MetricsMonitorResourceModeOrDefault(t, MetricsMonitorResourceModeAuto)
}
