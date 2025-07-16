//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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
	"bytes"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

const DefaultTokenSecretSize = 64

var secretTrimCharacters = map[byte]any{
	' ':  true,
	'\t': true,
	'\n': true,
	'\r': true,
}

func isTrimCharacter(char byte) bool {
	_, exists := secretTrimCharacters[char]
	return exists
}

func trimSecret(in []byte) []byte {
	return bytes.TrimFunc(in, func(r rune) bool {
		return isTrimCharacter(byte(r))
	})
}

func NewSecret(data []byte) Secret {
	return NewSecretWithSize(data, DefaultTokenSecretSize)
}

func NewSecretWithSize(data []byte, size int) Secret {
	data = trimSecret(data)

	if len(data) == 0 {
		return emptySecret{}
	}

	r := make([]byte, size)

	copy(r, data)

	return secret(r)
}

type secret []byte

func (s secret) Exists() bool {
	return true
}

func (s secret) Sign(method jwt.SigningMethod, claims Claims) (string, error) {
	token := jwt.NewWithClaims(method, jwt.MapClaims(claims))

	// Sign and get the complete encoded token as a string using the secret
	signedToken, err := token.SignedString([]byte(s))
	if err != nil {
		return "", errors.WithStack(err)
	}

	return signedToken, nil
}

func (s secret) Validate(token string) (Token, error) {
	return Validate(token, s)
}

func (s secret) Hash() string {
	return util.SHA256(s)
}
