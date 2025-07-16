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
	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewSecrets(secrets ...Secret) Secret {
	return Secrets(secrets)
}

type Secrets []Secret

func (s Secrets) Exists() bool {
	for _, k := range s {
		if k.Exists() {
			return true
		}
	}

	return false
}

func (s Secrets) Hash() string {
	return util.SHA256FromStringArray(util.FormatList(s, func(a Secret) string {
		return a.Hash()
	})...)
}

func (s Secrets) Sign(method jwt.SigningMethod, claims Claims) (string, error) {
	return "", errors.Errorf("secrets signing method not supported")
}

func (s Secrets) Validate(token string) (Token, error) {
	for _, secret := range s {
		if c, err := secret.Validate(token); err == nil {
			return c, nil
		} else {
			if !IsSignatureInvalidError(err) {
				return nil, err
			}
		}
	}

	return nil, jwt.ErrSignatureInvalid
}
