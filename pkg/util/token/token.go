//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	jg "github.com/golang-jwt/jwt"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	ClaimISS               = "iss"
	ClaimISSValue          = "arangodb"
	ClaimEXP               = "exp"
	ClaimIAT               = "iat"
	ClaimPreferredUsername = "preferred_username"
)

type Mod func(in Claims) Claims

func NewClaims() Claims {
	return Claims{}
}

type Claims jg.MapClaims

func (t Claims) With(mods ...Mod) Claims {
	q := t

	if q == nil {
		q = Claims{}
	}

	for _, mod := range mods {
		q = mod(q)
	}

	return q
}

func New(secret []byte, claims map[string]interface{}) (string, error) {
	token := jg.NewWithClaims(jg.SigningMethodHS256, jg.MapClaims(claims))

	// Sign and get the complete encoded token as a string using the secret
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return signedToken, nil
}
