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

package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cert"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func GenerateECDSASecret() ([]byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

func NewECDSAFromData(data []byte) (Secret, error) {
	keys, err := cert.LoadBytes(data)
	if err != nil {
		return nil, err
	}

	pKeys := util.FilterListType[*ecdsa.PrivateKey](keys)
	publicKeys := util.FilterListType[*ecdsa.PublicKey](keys)

	if len(pKeys) > 1 {
		return nil, errors.New("too many private keys provided")
	}

	return NewECDSASecret(util.First(pKeys...), publicKeys...)
}

func NewECDSASecret(key *ecdsa.PrivateKey, public ...*ecdsa.PublicKey) (Secret, error) {
	if key == nil && len(public) == 0 {
		// Return empty if nothing provided
		return EmptySecret(), nil
	} else if key == nil {
		// Validate only if provided
		validate, err := NewECDSAValidateSecretSet(public...)
		if err != nil {
			return nil, err
		}
		return Secrets(validate), nil
	} else {
		sign, err := NewECDSASignSecret(key)
		if err != nil {
			return nil, err
		}

		validate, err := NewECDSAValidateSecretSet(public...)
		if err != nil {
			return nil, err
		}

		return NewSecretSet(sign, validate...), nil
	}
}
