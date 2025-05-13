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
	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func Parse(token string, secret []byte) (Claims, error) {
	parsedToken, err := jwt.Parse(token,
		func(token *jwt.Token) (i interface{}, err error) {
			return secret, nil
		},
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}

	tokenClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.Errorf("Invalid token provided")
	}

	return Claims(tokenClaims), nil
}

func ParseWithAny(token string, secrets ...[]byte) (Claims, error) {
	for _, secret := range secrets {
		if c, err := Parse(token, secret); err != nil {
			if IsSignatureInvalidError(err) {
				continue
			}

			return nil, err
		} else {
			return c, nil
		}
	}

	return nil, jwt.ErrSignatureInvalid
}
