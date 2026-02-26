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
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func KeyToPublicPem(key any) (string, error) {
	switch key := key.(type) {
	case *ecdsa.PrivateKey:
		return PublicKeyToPem(&key.PublicKey)
	case *rsa.PrivateKey:
		return PublicKeyToPem(&key.PublicKey)
	case *ecdsa.PublicKey:
		return PublicKeyToPem(key)
	case *rsa.PublicKey:
		return PublicKeyToPem(key)
	}

	return "", errors.Errorf("private key type not supported")
}

func PublicKeyToPem(in any) (string, error) {
	// Marshal to PKIX (SubjectPublicKeyInfo)
	der, err := x509.MarshalPKIXPublicKey(in)
	if err != nil {
		return "", err
	}

	// Encode to PEM
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}

	return string(pem.EncodeToMemory(block)), nil
}
