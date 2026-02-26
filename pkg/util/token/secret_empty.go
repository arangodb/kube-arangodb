//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func EmptySecret() Secret {
	return emptySecret{}
}

type emptySecret struct{}

func (e emptySecret) PublicKey() []string {
	return nil
}

func (e emptySecret) KeyFunc(token *jwt.Token) (any, error) {
	return nil, noTokenFound{}
}

func (e emptySecret) Details(token string) (*string, []string, time.Duration, error) {
	return nil, []string{}, time.Duration(0), jwt.ErrSignatureInvalid
}

func (e emptySecret) SigningHash() string {
	return ""
}

func (e emptySecret) Hash() string {
	return ""
}

func (e emptySecret) Sign(claims Claims) (string, error) {
	return "", noTokenFound{}
}

func (e emptySecret) Method() jwt.SigningMethod {
	return jwt.SigningMethodNone
}

func (e emptySecret) Validate(token string) (Token, error) {
	return nil, jwt.ErrSignatureInvalid
}

func (e emptySecret) Exists() bool {
	return false
}
