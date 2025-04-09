//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	jwt "github.com/golang-jwt/jwt"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func newToken(in *jwt.Token) (Token, error) {
	tokenClaims, ok := in.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.Errorf("Invalid token provided")
	}

	if !in.Valid {
		return nil, &jwt.ValidationError{
			Inner:  jwt.ErrSignatureInvalid,
			Errors: 1,
		}
	}

	return token{
		claims: Claims(tokenClaims),
		valid:  in.Valid,
	}, nil
}

type token struct {
	claims Claims
	valid  bool
}

func (t token) Valid() bool {
	return t.valid
}

func (t token) Claims() Claims {
	return t.claims
}
