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
	"fmt"
	"net"
	"time"

	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
	"github.com/arangodb/k8s-operator/pkg/util/validation"
)

const (
	defaultTLSTTL = time.Hour * 2160 // About 3 month
)

// TLSSpec holds TLS specific configuration settings
type TLSSpec struct {
	CASecretName string        `json:"caSecretName,omitempty"`
	AltNames     []string      `json:"serverName,omitempty"`
	TTL          time.Duration `json:"ttl,omitempty"`
}

// IsSecure returns true when a CA secret has been set, false otherwise.
func (s TLSSpec) IsSecure() bool {
	return s.CASecretName != ""
}

// GetAltNames splits the list of AltNames into DNS names, IP addresses & email addresses.
// When an entry is not valid for any of those categories, an error is returned.
func (s TLSSpec) GetAltNames() (dnsNames, ipAddresses, emailAddresses []string, err error) {
	for _, name := range s.AltNames {
		if net.ParseIP(name) != nil {
			ipAddresses = append(ipAddresses, name)
		} else if validation.IsValidDNSName(name) {
			dnsNames = append(dnsNames, name)
		} else if validation.IsValidEmailAddress(name) {
			emailAddresses = append(emailAddresses, name)
		} else {
			return nil, nil, nil, maskAny(fmt.Errorf("'%s' is not a valid alternate name", name))
		}
	}
	return dnsNames, ipAddresses, emailAddresses, nil
}

// Validate the given spec
func (s TLSSpec) Validate() error {
	if err := k8sutil.ValidateOptionalResourceName(s.CASecretName); err != nil {
		return maskAny(err)
	}
	if _, _, _, err := s.GetAltNames(); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *TLSSpec) SetDefaults(defaultCASecretName string) {
	if s.CASecretName == "" {
		s.CASecretName = defaultCASecretName
	}
	if s.TTL == 0 {
		s.TTL = defaultTLSTTL
	}
}
