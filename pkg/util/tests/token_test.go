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

package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func Test_TokenSwitch(t *testing.T) {
	tm := NewTokenManager(t)

	jwt := GenerateJWTToken()
	key := GenerateECDSAP256Token(t)

	jwtToken := jwt.Sign(t)
	keyToken := key.Sign(t)

	t.Run("Try sign without token", func(t *testing.T) {
		err := tm.Error(utilToken.WithRelativeDuration(time.Second))
		require.Error(t, err)
		require.True(t, utilToken.IsNoTokenFoundError(err))
	})

	t.Run("Validate on empty folder", func(t *testing.T) {
		require.False(t, tm.Validate(t, jwtToken))
		require.False(t, tm.Validate(t, keyToken))
	})

	t.Run("Validate on jwt active", func(t *testing.T) {
		tm.Set(t, jwt, key)

		require.True(t, tm.Validate(t, jwtToken))
		require.True(t, tm.Validate(t, keyToken))
	})

	t.Run("Validate on ecdsa active", func(t *testing.T) {
		tm.Set(t, key, jwt)

		require.True(t, tm.Validate(t, jwtToken))
		require.True(t, tm.Validate(t, keyToken))
	})

	t.Run("Validate on only jwt active", func(t *testing.T) {
		tm.Set(t, jwt)

		require.True(t, tm.Validate(t, jwtToken))
		require.False(t, tm.Validate(t, keyToken))
	})

	t.Run("Validate on only ecdsa active", func(t *testing.T) {
		tm.Set(t, key)

		require.False(t, tm.Validate(t, jwtToken))
		require.True(t, tm.Validate(t, keyToken))
	})
}
