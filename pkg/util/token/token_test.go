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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func secret() []byte {
	d := make([]byte, 32)

	util.Rand().Read(d)

	return d
}

func sign(t *testing.T, secret []byte, mods ...Mod) string {
	token, err := New(secret, NewClaims().With(mods...))
	require.NoError(t, err)
	return token
}

func Test_TokenSign(t *testing.T) {
	t.Run("Signed properly token", func(t *testing.T) {
		s := secret()

		token := sign(t, s, WithCurrentIAT())

		claims, err := Parse(token, s)
		require.NoError(t, err)

		require.Contains(t, claims, ClaimIAT)
	})

	t.Run("Signed in future token", func(t *testing.T) {
		s := secret()

		token := sign(t, s, WithIAT(time.Now().Add(time.Hour)))

		_, err := Parse(token, s)
		require.EqualError(t, err, "Token used before issued")
	})

	t.Run("Expired", func(t *testing.T) {
		s := secret()

		token := sign(t, s, WithIAT(time.Now().Add(-time.Hour)), WithDuration(-time.Second))

		_, err := Parse(token, s)
		require.EqualError(t, err, "Token is expired")
	})

	t.Run("Invalid secret", func(t *testing.T) {
		s := secret()
		s2 := secret()

		token := sign(t, s, WithCurrentIAT())

		_, err := Parse(token, s2)
		require.EqualError(t, err, "signature is invalid")
		require.True(t, IsSignatureInvalidError(err))
	})

	t.Run("Signed properly token with first", func(t *testing.T) {
		s := secret()
		s2 := secret()

		token := sign(t, s, WithCurrentIAT())

		claims, err := ParseWithAny(token, s, s2)
		require.NoError(t, err)

		require.Contains(t, claims, ClaimIAT)
	})

	t.Run("Signed properly token with second", func(t *testing.T) {
		s := secret()
		s2 := secret()

		token := sign(t, s, WithCurrentIAT())

		claims, err := ParseWithAny(token, s2, s)
		require.NoError(t, err)

		require.Contains(t, claims, ClaimIAT)
	})

	t.Run("Without secrets", func(t *testing.T) {
		s := secret()

		token := sign(t, s, WithCurrentIAT())

		_, err := ParseWithAny(token)
		require.True(t, IsSignatureInvalidError(err))
	})

	t.Run("Expired with second", func(t *testing.T) {
		s := secret()
		s2 := secret()

		token := sign(t, s, WithIAT(time.Now().Add(-time.Hour)), WithDuration(-time.Second))

		_, err := ParseWithAny(token, s2, s)
		require.EqualError(t, err, "Token is expired")
	})
}
