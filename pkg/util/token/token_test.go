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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func testSecretToken() []byte {
	return testSecretTokenSized(64)
}

func testSecretTokenSized(size int) []byte {
	var z = make([]byte, size)
	for id := range z {
		z[id] = byte('A' + util.Rand().Intn('Z'-'A'+1))
	}
	return z
}

func Test_TokenSign(t *testing.T) {
	t.Run("Signed properly token", func(t *testing.T) {
		s := NewSecret(testSecretToken())

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s)
		require.NoError(t, err)

		claims, err := s.Validate(token)
		require.NoError(t, err)

		require.Contains(t, claims.Claims(), ClaimIAT)
	})

	t.Run("Signed in future token", func(t *testing.T) {
		s := NewSecret(testSecretToken())

		token, err := NewClaims().With(WithIAT(time.Now().Add(time.Hour))).Sign(s)
		require.NoError(t, err)

		_, err = s.Validate(token)
		require.EqualError(t, err, "token has invalid claims: token used before issued")
	})

	t.Run("Expired", func(t *testing.T) {
		s := NewSecret(testSecretToken())

		token, err := NewClaims().With(WithIAT(time.Now().Add(-time.Hour)), WithDuration(-time.Second)).Sign(s)
		require.NoError(t, err)

		_, err = s.Validate(token)
		require.EqualError(t, err, "token has invalid claims: token is expired")
	})

	t.Run("Invalid secret", func(t *testing.T) {
		s := NewSecret(testSecretToken())
		s2 := NewSecret(testSecretToken())

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s)
		require.NoError(t, err)

		_, err = s2.Validate(token)
		require.EqualError(t, err, "token signature is invalid: signature is invalid")
		require.True(t, IsSignatureInvalidError(err))
	})

	t.Run("Signed properly token with first", func(t *testing.T) {
		s := NewSecret(testSecretToken())
		s2 := NewSecret(testSecretToken())

		sm := NewSecretSet(s, s2)

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s)
		require.NoError(t, err)

		claims, err := sm.Validate(token)
		require.NoError(t, err)

		require.Contains(t, claims.Claims(), ClaimIAT)
	})

	t.Run("Signed properly token with second", func(t *testing.T) {
		s := NewSecret(testSecretToken())
		s2 := NewSecret(testSecretToken())

		sm := NewSecretSet(s, s2)

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s2)
		require.NoError(t, err)

		claims, err := sm.Validate(token)
		require.NoError(t, err)

		require.Contains(t, claims.Claims(), ClaimIAT)
	})

	t.Run("Without secrets", func(t *testing.T) {
		s := NewSecret(testSecretToken())
		ns := NewSecrets()

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s)
		require.NoError(t, err)

		_, err = ns.Validate(token)
		require.True(t, IsSignatureInvalidError(err))
	})

	t.Run("Expired with second", func(t *testing.T) {
		s := NewSecret(testSecretToken())
		s2 := NewSecret(testSecretToken())

		ns := NewSecretSet(s, s2)

		token, err := NewClaims().With(WithIAT(time.Now().Add(-time.Hour)), WithDuration(-time.Second)).Sign(s2)
		require.NoError(t, err)

		_, err = ns.Validate(token)
		require.EqualError(t, err, "token has invalid claims: token is expired")
	})

	t.Run("Ensure token gets trimmed", func(t *testing.T) {
		b := testSecretTokenSized(128)
		s := NewSecret(b)
		s2 := NewSecret(b[:64])

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s)
		require.NoError(t, err)

		claims, err := s2.Validate(token)
		require.NoError(t, err)

		require.Contains(t, claims.Claims(), ClaimIAT)
	})

	t.Run("Ensure token gets filled", func(t *testing.T) {
		b := testSecretTokenSized(16)
		bs := make([]byte, DefaultTokenSecretSize)
		copy(bs, b)
		s := NewSecret(b)
		s2 := NewSecret(bs)

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s)
		require.NoError(t, err)

		claims, err := s2.Validate(token)
		require.NoError(t, err)

		require.Contains(t, claims.Claims(), ClaimIAT)
	})

	t.Run("Ensure token gets removed prefix", func(t *testing.T) {
		b := testSecretTokenSized(16)
		bs := make([]byte, DefaultTokenSecretSize)
		bs[0] = ' '
		copy(bs[1:], b)
		s := NewSecret(b)
		s2 := NewSecret(bs)

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s)
		require.NoError(t, err)

		claims, err := s2.Validate(token)
		require.NoError(t, err)

		require.Contains(t, claims.Claims(), ClaimIAT)
	})

	t.Run("Ensure token gets removed postfix", func(t *testing.T) {
		b := testSecretTokenSized(16)
		bs := make([]byte, 17)
		bs[16] = ' '
		copy(bs, b)
		s := NewSecret(b)
		s2 := NewSecret(bs)

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s)
		require.NoError(t, err)

		claims, err := s2.Validate(token)
		require.NoError(t, err)

		require.Contains(t, claims.Claims(), ClaimIAT)
	})

	t.Run("Ensure token gets trimmed", func(t *testing.T) {
		b := testSecretTokenSized(16)
		bs := make([]byte, 18)
		bs[17] = ' '
		bs[0] = ' '
		copy(bs[1:], b)
		s := NewSecret(b)
		s2 := NewSecret(bs)

		token, err := NewClaims().With(WithCurrentIAT()).Sign(s)
		require.NoError(t, err)

		claims, err := s2.Validate(token)
		require.NoError(t, err)

		require.Contains(t, claims.Claims(), ClaimIAT)
	})
}
