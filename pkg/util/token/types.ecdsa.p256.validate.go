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
	"crypto/x509"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewECDSAValidateSecretSet(keys ...*ecdsa.PublicKey) ([]Secret, error) {
	r := make([]Secret, len(keys))

	for id := range keys {
		s, err := NewECDSAValidateSecret(keys[id])
		if err != nil {
			return nil, err
		}
		r[id] = s
	}

	return r, nil
}

func NewECDSAValidateSecret(key *ecdsa.PublicKey) (Secret, error) {
	if key == nil {
		return EmptySecret(), nil
	}

	data, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}

	return ecdsaValidateSecret{
		key:  key,
		hash: util.SHA256(data),
	}, nil
}

type ecdsaValidateSecret struct {
	key  *ecdsa.PublicKey
	hash string
}

func (e ecdsaValidateSecret) SigningHash() string {
	return ""
}

func (e ecdsaValidateSecret) Hash() string {
	return e.hash
}

func (e ecdsaValidateSecret) Sign(claims Claims) (string, error) {
	return "", errors.Errorf("ECDSA p256 validate secret is read-only and cannot sign new tokens")
}

func (e ecdsaValidateSecret) Validate(token string) (Token, error) {
	return Validate(token, e)
}

func (e ecdsaValidateSecret) KeyFunc(token *jwt.Token) (any, error) {
	return e.key, nil
}

func (e ecdsaValidateSecret) Method() jwt.SigningMethod {
	return jwt.SigningMethodES256
}

func (e ecdsaValidateSecret) Details(token string) (*string, []string, time.Duration, error) {
	return extractTokenDetails(e, token)
}

func (e ecdsaValidateSecret) Exists() bool {
	return true
}
