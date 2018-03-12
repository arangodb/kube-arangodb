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

package certificates

import (
	"crypto/tls"
	"crypto/x509"
	"time"
)

// GetCertificateExpirationDate returns the expiration date of the TLS certificate
// found in the given config.
// Returns: ExpirationDate, FoundExpirationDate
func GetCertificateExpirationDate(config *tls.Config) (time.Time, bool) {
	if config == nil || len(config.Certificates) == 0 {
		return time.Time{}, false
	}
	var expDate time.Time
	found := false
	for _, raw := range config.Certificates[0].Certificate {
		if c, err := x509.ParseCertificate(raw); err == nil {
			d := c.NotAfter
			if !found || d.Before(expDate) {
				expDate = d
			}
			found = true
		}
	}
	return expDate, found
}
