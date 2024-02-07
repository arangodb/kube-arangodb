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
	"time"

	jg "github.com/golang-jwt/jwt"
)

var defaultTokenClaims = jg.MapClaims{
	ClaimISS: ClaimISSValue,
}

func WithDefaultClaims() Mod {
	return func(in Claims) Claims {
		for k, v := range defaultTokenClaims {
			if _, ok := in[k]; !ok {
				in[k] = v
			}
		}

		return in
	}
}

func WithUsername(username string) Mod {
	return func(in Claims) Claims {
		in[ClaimPreferredUsername] = username
		return in
	}
}

func WithCurrentIAT() Mod {
	return func(in Claims) Claims {
		in[ClaimIAT] = time.Now().Unix()
		return in
	}
}

func WithIAT(time time.Time) Mod {
	return func(in Claims) Claims {
		in[ClaimIAT] = time.Unix()
		return in
	}
}

func WithDuration(dur time.Duration) Mod {
	return func(in Claims) Claims {
		in[ClaimEXP] = time.Now().Add(dur).Unix()
		return in
	}
}

func WithExp(time time.Time) Mod {
	return func(in Claims) Claims {
		in[ClaimEXP] = time.Unix()
		return in
	}
}
