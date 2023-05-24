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
// Adam Janikowski
//

package v1

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// MetricsAuthenticationSpec contains spec for authentication with arangodb
type MetricsAuthenticationSpec struct {
	// JWTTokenSecretName contains the name of the JWT kubernetes secret used for authentication
	JWTTokenSecretName *string `json:"jwtTokenSecretName,omitempty"`
}

// MetricsMode defines mode for metrics exporter
type MetricsMode string

func (m MetricsMode) New() *MetricsMode {
	return &m
}

// GetMetricsEndpoint
// Deprecated
func (m MetricsMode) GetMetricsEndpoint() string {
	switch m {
	case MetricsModeInternal:
		return shared.ArangoExporterInternalEndpoint
	default:
		return shared.ArangoExporterDefaultEndpoint
	}
}

const (
	// MetricsModeExporter starts sidecar container with
	// Deprecated
	MetricsModeExporter MetricsMode = "exporter"
	// MetricsModeSidecar behaves exactly the same as MetricsModeExporter
	// Deprecated
	MetricsModeSidecar MetricsMode = "sidecar"
	// MetricsModeInternal exposes metrics using ArangoD endpoint
	// Deprecated
	MetricsModeInternal MetricsMode = "internal"
)

func (m *MetricsMode) Get() MetricsMode {
	if m == nil {
		return MetricsModeExporter
	}

	return *m
}

// MetricsSpec contains spec for arangodb exporter
type MetricsSpec struct {
	Enabled *bool `json:"enabled,omitempty"`
	// deprecated
	Image          *string                   `json:"image,omitempty"`
	Authentication MetricsAuthenticationSpec `json:"authentication,omitempty"`
	Resources      core.ResourceRequirements `json:"resources,omitempty"`
	// deprecated
	Mode *MetricsMode `json:"mode,omitempty"`
	TLS  *bool        `json:"tls,omitempty"`

	ServiceMonitor *MetricsServiceMonitorSpec `json:"serviceMonitor,omitempty"`

	Port *uint16 `json:"port,omitempty"`
}

func (s *MetricsSpec) IsTLS() bool {
	if s == nil || s.TLS == nil {
		return true
	}

	return *s.TLS
}

func (s *MetricsSpec) GetPort() uint16 {
	if s == nil || s.Port == nil {
		return shared.ArangoExporterPort
	}

	return *s.Port
}

// IsEnabled returns whether metrics are enabled or not
func (s *MetricsSpec) IsEnabled() bool {
	return util.TypeOrDefault[bool](s.Enabled, false)
}

// HasImage returns whether a image was specified or not
// Deprecated
func (s *MetricsSpec) HasImage() bool {
	return s.Image != nil
}

// GetImage returns the Image or empty string
// Deprecated
func (s *MetricsSpec) GetImage() string {
	return util.TypeOrDefault[string](s.Image)
}

// SetDefaults sets default values
func (s *MetricsSpec) SetDefaults(defaultTokenName string, isAuthenticated bool) {
	if s.Enabled == nil {
		s.Enabled = util.NewType[bool](false)
	}
	if s.GetJWTTokenSecretName() == "" {
		s.Authentication.JWTTokenSecretName = util.NewType[string](defaultTokenName)
	}
}

// GetJWTTokenSecretName returns the token secret name or empty string
func (s *MetricsSpec) GetJWTTokenSecretName() string {
	return util.TypeOrDefault[string](s.Authentication.JWTTokenSecretName)
}

// HasJWTTokenSecretName returns true if a secret name was specified
func (s *MetricsSpec) HasJWTTokenSecretName() bool {
	return s.Authentication.JWTTokenSecretName != nil
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *MetricsSpec) SetDefaultsFrom(source MetricsSpec) {
	if s.Enabled == nil {
		s.Enabled = util.NewTypeOrNil[bool](source.Enabled)
	}
	if s.Image == nil {
		s.Image = util.NewTypeOrNil[string](source.Image)
	}
	if s.Authentication.JWTTokenSecretName == nil {
		s.Authentication.JWTTokenSecretName = util.NewTypeOrNil[string](source.Authentication.JWTTokenSecretName)
	}
	setStorageDefaultsFromResourceList(&s.Resources.Limits, source.Resources.Limits)
	setStorageDefaultsFromResourceList(&s.Resources.Requests, source.Resources.Requests)
}

// Validate the given spec
func (s *MetricsSpec) Validate() error {

	if s.HasJWTTokenSecretName() {
		if err := shared.ValidateResourceName(s.GetJWTTokenSecretName()); err != nil {
			return err
		}
	}

	return nil
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
func (s MetricsSpec) ResetImmutableFields(fieldPrefix string, target *MetricsSpec) []string {
	return nil
}
