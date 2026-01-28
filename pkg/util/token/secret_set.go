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
	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewSecretSet(main Secret, secondary ...Secret) Secret {
	return secretSet{
		main:      main,
		secondary: Secrets(secondary),
	}
}

type secretSet struct {
	main Secret

	secondary Secret
}

func (s secretSet) SigningHash() string {
	return s.main.SigningHash()
}

func (s secretSet) Exists() bool {
	return s.main.Exists() || s.secondary.Exists()
}

func (s secretSet) Hash() string {
	return util.SHA256FromStringArray(s.main.Hash(), s.secondary.Hash())
}

func (s secretSet) Sign(method jwt.SigningMethod, claims Claims) (string, error) {
	return s.main.Sign(method, claims)
}

func (s secretSet) Validate(token string) (Token, error) {
	if c, err := s.main.Validate(token); err == nil {
		return c, nil
	}

	return s.secondary.Validate(token)
}
