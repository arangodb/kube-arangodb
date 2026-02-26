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
	"github.com/arangodb/kube-arangodb/pkg/util/cert"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewECDSASignSecret(key *ecdsa.PrivateKey) (Secret, error) {
	if key == nil {
		return EmptySecret(), nil
	}

	data, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}

	return ecdsaSigningSecret{
		key:  key,
		hash: util.SHA256(data),
	}, nil
}

type ecdsaSigningSecret struct {
	key  *ecdsa.PrivateKey
	hash string
}

func (e ecdsaSigningSecret) PublicKey() []string {
	res, err := cert.KeyToPublicPem(e.key)
	if err != nil {
		return nil
	}

	return []string{res}
}

func (e ecdsaSigningSecret) KeyFunc(token *jwt.Token) (any, error) {
	return &e.key.PublicKey, nil
}

func (e ecdsaSigningSecret) Method() jwt.SigningMethod {
	return jwt.SigningMethodES256
}

func (e ecdsaSigningSecret) SigningHash() string {
	return e.hash
}

func (e ecdsaSigningSecret) Hash() string {
	return e.hash
}

func (e ecdsaSigningSecret) Sign(claims Claims) (string, error) {
	token := jwt.NewWithClaims(e.Method(), jwt.MapClaims(claims))

	// Sign and get the complete encoded token as a string using the secret
	signedToken, err := token.SignedString(e.key)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return signedToken, nil
}

func (e ecdsaSigningSecret) Validate(token string) (Token, error) {
	return Validate(token, e)
}

func (e ecdsaSigningSecret) Details(token string) (*string, []string, time.Duration, error) {
	return extractTokenDetails(e, token)
}

func (e ecdsaSigningSecret) Exists() bool {
	return true
}
