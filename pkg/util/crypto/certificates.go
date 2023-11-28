//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package crypto

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
)

type Certificates []*x509.Certificate

func (c Certificates) Contains(cert *x509.Certificate) bool {
	for _, localCert := range c {
		if !localCert.Equal(cert) {
			return false
		}
	}

	return true
}

func (c Certificates) ContainsAll(certs Certificates) bool {
	if len(certs) == 0 {
		return true
	}

	for _, cert := range certs {
		if !c.Contains(cert) {
			return false
		}
	}

	return true
}

func (c Certificates) ToPem() ([]byte, error) {
	bytes := bytes.NewBuffer([]byte{})

	for _, cert := range c {
		if err := pem.Encode(bytes, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
			return nil, err
		}
	}

	return bytes.Bytes(), nil
}

func (c Certificates) AsCertPool() *x509.CertPool {
	cp := x509.NewCertPool()

	for _, cert := range c {
		cp.AddCert(cert)
	}

	return cp
}
