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
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewHelper(t *testing.T) {
	ctx := context.Background()

	t.Run("Returns the rendered token", func(t *testing.T) {
		var calls int
		h := NewHelper[string](func(ctx context.Context, in string) (Token, error) {
			calls++
			return Token("tok-" + in), nil
		})

		v, err := h.Get(ctx, "user")
		require.NoError(t, err)
		require.Equal(t, Token("tok-user"), v)

		// Second call for the same key is served from cache.
		v, err = h.Get(ctx, "user")
		require.NoError(t, err)
		require.Equal(t, Token("tok-user"), v)
		require.Equal(t, 1, calls)
	})

	t.Run("Propagates the error", func(t *testing.T) {
		h := NewHelper[string](func(ctx context.Context, in string) (Token, error) {
			return "", fmt.Errorf("boom")
		})

		_, err := h.Get(ctx, "user")
		require.EqualError(t, err, "boom")
	})
}

type helperIface struct{}

func (helperIface) Token(ctx context.Context, in string) (Token, error) {
	return Token("iface-" + in), nil
}

func Test_NewHelperInterface(t *testing.T) {
	h := NewHelperInterface[string](helperIface{})

	v, err := h.Get(context.Background(), "user")
	require.NoError(t, err)
	require.Equal(t, Token("iface-user"), v)
}
