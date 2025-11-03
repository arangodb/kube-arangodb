//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	LicenseExpirationGraceRatio         = 0.9
	DefaultLicenseExpirationGracePeriod = 3 * 24 * time.Hour
	DefaultLicenseTTL                   = 14 * 24 * time.Hour
)

type LicenseMode string

const (
	LicenseModeDefault              = LicenseModeDiscover
	LicenseModeDiscover LicenseMode = "discover"
	LicenseModeKey      LicenseMode = "key"
	LicenseModeAPI      LicenseMode = "api"
)

func (l *LicenseMode) Get() LicenseMode {
	return util.OptionalType(l, LicenseModeDefault)
}

// LicenseSpec holds the license related information
type LicenseSpec struct {
	// SecretName setting specifies the name of a kubernetes `Secret` that contains
	// the license key token or master key used for enterprise images. This value is not used for
	// the Community Edition.
	SecretName *string `json:"secretName,omitempty"`

	// Mode Defines the mode of license
	// +doc/default: discover
	// +doc/enum: discover|Discovers the LicenseMode based on the keys
	// +doc/enum: key|Use License Key mechanism
	// +doc/enum: master|Use License Master Key mechanism
	Mode *LicenseMode `json:"mode,omitempty"`

	// TTL Sets the requested License TTL
	// +doc/default: 336h
	TTL *meta.Duration `json:"ttl,omitempty"`

	// ExpirationGracePeriod defines the expiration grace period for the license
	// +doc/default: 72h
	ExpirationGracePeriod *meta.Duration `json:"expirationGracePeriod,omitempty"`

	// Telemetry defines if telemetry is collected
	// +doc/default: true
	Telemetry *bool `json:"telemetry,omitempty"`
}

// HasSecretName returns true if a license key secret name was set
func (s LicenseSpec) HasSecretName() bool {
	return s.SecretName != nil
}

// GetSecretName returns the license key if set. Empty string otherwise.
func (s LicenseSpec) GetSecretName() string {
	return util.TypeOrDefault[string](s.SecretName)
}

// GetTTL returns the license TTL
func (s LicenseSpec) GetTTL() time.Duration {
	if s.TTL == nil {
		return DefaultLicenseTTL
	}
	return s.TTL.Duration
}

// GetTelemetry returns the license Telemetry
func (s LicenseSpec) GetTelemetry() bool {
	return util.OptionalType(s.Telemetry, true)
}

// GetExpirationGracePeriod returns the expiration period
func (s LicenseSpec) GetExpirationGracePeriod() time.Duration {
	if s.ExpirationGracePeriod == nil {
		return DefaultLicenseExpirationGracePeriod
	}
	return s.ExpirationGracePeriod.Duration
}

// Validate validates the LicenseSpec
func (s LicenseSpec) Validate() error {
	if !s.HasSecretName() {
		return nil
	}
	return shared.WithErrors(
		// Secret
		shared.PrefixResourceErrorFunc("secretName", func() error {
			return shared.ValidateResourceName(s.GetSecretName())
		}),
		// Expiration
		shared.PrefixResourceErrorFunc("expirationGracePeriod", func() error {
			if s.GetExpirationGracePeriod() <= 0 {
				return errors.Errorf("Expiration grace period must be greater than zero")
			}

			if s.GetExpirationGracePeriod() >= s.GetTTL() {
				return errors.Errorf("Expiration grace period must be less than TTL")
			}

			return nil
		}),
		// TTL
		shared.PrefixResourceErrorFunc("ttl", func() error {
			if s.GetTTL() <= 0 {
				return errors.Errorf("TTL must be greater than zero")
			}

			return nil
		}),
	)
}

// SetDefaultsFrom fills all values not set in s with values from other
func (s *LicenseSpec) SetDefaultsFrom(other LicenseSpec) {
	if !s.HasSecretName() {
		s.SecretName = util.NewTypeOrNil(other.SecretName)
	}
	s.TTL = util.NewTypeOrNil(other.TTL)
	s.ExpirationGracePeriod = util.NewTypeOrNil(other.ExpirationGracePeriod)
}
