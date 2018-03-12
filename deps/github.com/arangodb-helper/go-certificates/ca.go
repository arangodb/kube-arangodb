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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

type CA struct {
	Certificate []*x509.Certificate
	PrivateKey  interface{}
}

// LoadCAFromPEM parses the given certificate & key into a CA instance.
func LoadCAFromPEM(cert, key string) (CA, error) {
	certs, privKey, err := LoadFromPEM(cert, key)
	if err != nil {
		return CA{}, maskAny(err)
	}
	return CA{
		Certificate: certs,
		PrivateKey:  privKey,
	}, nil
}

// LoadFromPEM parses the given certificate & key into a certificate slice & private key.
func LoadFromPEM(cert, key string) ([]*x509.Certificate, interface{}, error) {
	var certs []*x509.Certificate

	// Parse certificate
	pemCerts := []byte(cert)
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		c, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, nil, maskAny(err)
		}

		certs = append(certs, c)
	}
	if len(certs) == 0 {
		return nil, nil, maskAny(fmt.Errorf("No CERTIFICATE's found in '%s'", cert))
	}

	// Parse key
	pemKey := []byte(key)
	var privKey interface{}
	for len(pemKey) > 0 {
		var block *pem.Block
		block, pemKey = pem.Decode(pemKey)
		if block == nil {
			break
		}

		if block.Type == "PRIVATE KEY" || strings.HasSuffix(block.Type, " PRIVATE KEY") {
			if privKey == nil {
				var err error
				privKey, err = parsePrivateKey(block.Bytes)
				if err != nil {
					return nil, nil, maskAny(err)
				}
			}
		}
	}
	if privKey == nil {
		return nil, nil, maskAny(fmt.Errorf("No PRIVATE KEY found in '%s'", key))
	}

	return certs, privKey, nil
}
