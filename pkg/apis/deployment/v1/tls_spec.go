//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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
	"net"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/validation"
)

type TLSRotateMode string

func (t *TLSRotateMode) Get() TLSRotateMode {
	if t == nil {
		return TLSRotateModeInPlace
	}

	return *t
}

func (t TLSRotateMode) New() *TLSRotateMode {
	return &t
}

const (
	TLSRotateModeInPlace  TLSRotateMode = "inplace"
	TLSRotateModeRecreate TLSRotateMode = "recreate"
)

const (
	DefaultTLSTTL = Duration("2610h") // About 3 month
)

// TLSSpec holds TLS specific configuration settings
type TLSSpec struct {
	// CASecretName  setting specifies the name of a kubernetes `Secret` that contains
	// a standard CA certificate + private key used to sign certificates for individual
	// ArangoDB servers.
	// When no name is specified, it defaults to `<deployment-name>-ca`.
	// To disable authentication, set this value to `None`.
	// If you specify a name of a `Secret` that does not exist, a self-signed CA certificate + key is created
	// and stored in a `Secret` with given name.
	// The specified `Secret`, must contain the following data fields:
	// - `ca.crt` PEM encoded public key of the CA certificate
	// - `ca.key` PEM encoded private key of the CA certificate
	CASecretName *string `json:"caSecretName,omitempty"`
	// AltNames setting specifies a list of alternate names that will be added to all generated
	// certificates. These names can be DNS names or email addresses.
	// The default value is empty.
	// +doc/type: []string
	AltNames []string `json:"altNames,omitempty"`
	// TTL setting specifies the time to live of all generated server certificates.
	// When the server certificate is about to expire, it will be automatically replaced
	// by a new one and the affected server will be restarted.
	// Note: The time to live of the CA certificate (when created automatically)
	// will be set to 10 years.
	// +doc/default: "2160h" (about 3 months)
	TTL  *Duration      `json:"ttl,omitempty"`
	SNI  *TLSSNISpec    `json:"sni,omitempty"`
	Mode *TLSRotateMode `json:"mode,omitempty"`
}

const (
	// CASecretNameDisabled is the value of CASecretName to use for disabling authentication.
	CASecretNameDisabled = "None"
)

// GetCASecretName returns the value of caSecretName.
func (s TLSSpec) GetCASecretName() string {
	return util.TypeOrDefault[string](s.CASecretName)
}

// GetAltNames returns the value of altNames.
func (s TLSSpec) GetAltNames() []string {
	return s.AltNames
}

// GetTTL returns the value of ttl.
func (s TLSSpec) GetTTL() Duration {
	return DurationOrDefault(s.TTL)
}

func (a TLSSpec) GetSNI() TLSSNISpec {
	if a.SNI == nil {
		return TLSSNISpec{}
	}

	return *a.SNI
}

// IsSecure returns true when a CA secret has been set, false otherwise.
func (s TLSSpec) IsSecure() bool {
	return s.GetCASecretName() != CASecretNameDisabled
}

// GetParsedAltNames splits the list of AltNames into DNS names, IP addresses & email addresses.
// When an entry is not valid for any of those categories, an error is returned.
func (s TLSSpec) GetParsedAltNames() (dnsNames, ipAddresses, emailAddresses []string, err error) {
	for _, name := range s.GetAltNames() {
		if net.ParseIP(name) != nil {
			ipAddresses = append(ipAddresses, name)
		} else if validation.IsValidDNSName(name) {
			dnsNames = append(dnsNames, name)
		} else if validation.IsValidEmailAddress(name) {
			emailAddresses = append(emailAddresses, name)
		} else {
			return nil, nil, nil, errors.WithStack(errors.Errorf("'%s' is not a valid alternate name", name))
		}
	}
	return dnsNames, ipAddresses, emailAddresses, nil
}

// Validate the given spec
func (s TLSSpec) Validate() error {
	if s.IsSecure() {
		if err := shared.ValidateResourceName(s.GetCASecretName()); err != nil {
			return errors.WithStack(err)
		}
		if _, _, _, err := s.GetParsedAltNames(); err != nil {
			return errors.WithStack(err)
		}
		if err := s.GetTTL().Validate(); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *TLSSpec) SetDefaults(defaultCASecretName string) {
	if s.GetCASecretName() == "" {
		// Note that we don't check for nil here, since even a specified, but empty
		// string should result in the default value.
		s.CASecretName = util.NewType[string](defaultCASecretName)
	}
	if s.GetTTL() == "" {
		// Note that we don't check for nil here, since even a specified, but zero
		// should result in the default value.
		s.TTL = NewDuration(DefaultTLSTTL)
	}
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *TLSSpec) SetDefaultsFrom(source TLSSpec) {
	if s.CASecretName == nil {
		s.CASecretName = util.NewTypeOrNil[string](source.CASecretName)
	}
	if s.AltNames == nil {
		s.AltNames = source.AltNames
	}
	if s.TTL == nil {
		s.TTL = NewDurationOrNil(source.TTL)
	}
	if s.SNI == nil {
		s.SNI = source.SNI.DeepCopy()
	}
}
