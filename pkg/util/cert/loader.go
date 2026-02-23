//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package cert

import (
	"crypto/x509"
	"encoding/pem"
)

func LoadBytes(data []byte) ([]any, error) {
	var keys []any

	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break // no more PEM blocks
		}

		switch block.Type {

		case "PUBLIC KEY":
			pub, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			keys = append(keys, pub)

		case "PRIVATE KEY": // PKCS#8
			keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			keys = append(keys, keyAny)

		case "EC PRIVATE KEY":
			priv, err := x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			keys = append(keys, priv)

		case "RSA PRIVATE KEY":
			priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			keys = append(keys, priv)

		default:
			// ignore: CERTIFICATE, EC PARAMETERS, etc.
		}
	}

	return keys, nil
}
