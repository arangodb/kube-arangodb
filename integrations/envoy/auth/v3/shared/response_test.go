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

package shared

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ResponseAuth_Hash(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var a *ResponseAuth
		require.Equal(t, "", a.Hash())
	})

	t.Run("Deterministic", func(t *testing.T) {
		a := &ResponseAuth{User: "user", Groups: []string{"a", "b"}}
		require.Equal(t, a.Hash(), a.Hash())
	})

	t.Run("Group order does not affect the hash", func(t *testing.T) {
		a := &ResponseAuth{User: "user", Groups: []string{"a", "b", "c"}}
		b := &ResponseAuth{User: "user", Groups: []string{"c", "a", "b"}}
		require.Equal(t, a.Hash(), b.Hash())
	})

	t.Run("Different user changes the hash", func(t *testing.T) {
		a := &ResponseAuth{User: "user1", Groups: []string{"a"}}
		b := &ResponseAuth{User: "user2", Groups: []string{"a"}}
		require.NotEqual(t, a.Hash(), b.Hash())
	})

	t.Run("Different groups change the hash", func(t *testing.T) {
		a := &ResponseAuth{User: "user", Groups: []string{"a"}}
		b := &ResponseAuth{User: "user", Groups: []string{"a", "b"}}
		require.NotEqual(t, a.Hash(), b.Hash())
	})

	// Regression: Hash must not mutate the caller's Groups slice, which may be shared with the
	// authentication cache entry.
	t.Run("Does not mutate the input slice", func(t *testing.T) {
		groups := []string{"c", "a", "b"}
		a := &ResponseAuth{User: "user", Groups: groups}

		a.Hash()

		require.Equal(t, []string{"c", "a", "b"}, groups)
		require.Equal(t, []string{"c", "a", "b"}, a.Groups)
	})
}

func Test_Response_Authenticated(t *testing.T) {
	require.False(t, Response{}.Authenticated())
	require.True(t, Response{User: &ResponseAuth{User: "user"}}.Authenticated())
}

func Test_Response_AsResponse(t *testing.T) {
	r := Response{}
	resp := r.AsResponse()
	require.NotNil(t, resp)

	ok := resp.GetOkResponse()
	require.NotNil(t, ok)
	require.Empty(t, ok.GetHeaders())
	require.Empty(t, ok.GetResponseHeadersToAdd())
}
