//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Validate(t string, secret Secret) (Token, error) {
	token, err := jwt.Parse(t, secret.KeyFunc,
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{secret.Method().Alg()}),
	)
	if err != nil {
		return nil, err
	}

	return newToken(token)
}

func newToken(in *jwt.Token) (Token, error) {
	tokenClaims, ok := in.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.Errorf("Invalid token provided")
	}

	if !in.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return token{
		claims: Claims(tokenClaims),
	}, nil
}

type token struct {
	claims Claims
}

func (t token) Claims() Claims {
	return t.claims
}
