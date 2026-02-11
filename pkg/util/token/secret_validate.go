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
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func ValidateOnly(in Secret) Secret {
	return validateOnly{secret: in}
}

type validateOnly struct {
	secret Secret
}

func (v validateOnly) Details(token string) (*string, []string, time.Duration, error) {
	return extractTokenDetails(v, token)
}

func (v validateOnly) SigningHash() string {
	return v.secret.SigningHash()
}

func (v validateOnly) Hash() string {
	return v.secret.Hash()
}

func (v validateOnly) Sign(method jwt.SigningMethod, claims Claims) (string, error) {
	return "", errors.Errorf("Secret only allows validation")
}

func (v validateOnly) Validate(token string) (Token, error) {
	return v.secret.Validate(token)
}

func (v validateOnly) Exists() bool {
	return v.secret.Exists()
}
