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
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

var defaultTokenClaims = jwt.MapClaims{
	ClaimISS: ClaimISSValue,
}

func WithDefaultClaims() util.ModR[Claims] {
	return func(in Claims) Claims {
		for k, v := range defaultTokenClaims {
			if _, ok := in[k]; !ok {
				in[k] = v
			}
		}

		return in
	}
}

func WithUsername(username string) util.ModR[Claims] {
	return WithKey(ClaimPreferredUsername, username)
}

func WithCurrentIAT() util.ModR[Claims] {
	return WithIAT(time.Now())
}

func WithIAT(time time.Time) util.ModR[Claims] {
	return WithKey(ClaimIAT, time.Unix())
}

func WithDuration(dur time.Duration) util.ModR[Claims] {
	return WithExp(time.Now().Add(dur))
}

func WithExp(time time.Time) util.ModR[Claims] {
	return WithKey(ClaimEXP, time.Unix())
}

func WithServerID(id string) util.ModR[Claims] {
	return WithKey(ClaimServerID, id)
}

func WithAllowedPaths(paths ...string) util.ModR[Claims] {
	if len(paths) == 0 {
		return emptyClaimsMod
	}
	return WithKey(ClaimAllowedPaths, paths)
}

func emptyClaimsMod(in Claims) Claims {
	return in
}

func WithKey(key string, value interface{}) util.ModR[Claims] {
	return func(in Claims) Claims {
		in[key] = value
		return in
	}
}

func WithRoles(roles ...string) util.ModR[Claims] {
	return func(in Claims) Claims {
		in[ClaimRoles] = roles
		return in
	}
}
