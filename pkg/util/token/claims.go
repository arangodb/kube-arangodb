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

func NewClaims() Claims {
	return Claims{}.With(WithDefaultClaims())
}

type Claims jwt.MapClaims

func (t Claims) With(mods ...util.ModR[Claims]) Claims {
	q := t

	if q == nil {
		q = Claims{}
	}

	return util.ApplyModsR(q, mods...)
}

func (t Claims) Sign(secret Secret) (string, error) {
	return secret.Sign(t)
}
