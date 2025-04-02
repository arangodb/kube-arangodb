//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Cache_TTL(t *testing.T) {
	var z int

	c := NewCache[string, int](func(ctx context.Context, in string) (int, error) {
		q := z
		z++

		return q, nil
	}, 100*time.Millisecond)

	require.Equal(t, 0, tests.NoError[int](t)(c.Get(context.Background(), "k1")))
	require.Equal(t, 1, z)
	require.Equal(t, 0, tests.NoError[int](t)(c.Get(context.Background(), "k1")))
	require.Equal(t, 0, tests.NoError[int](t)(c.Get(context.Background(), "k1")))
	require.Equal(t, 0, tests.NoError[int](t)(c.Get(context.Background(), "k1")))
	require.Equal(t, 0, tests.NoError[int](t)(c.Get(context.Background(), "k1")))
	require.Equal(t, 1, tests.NoError[int](t)(c.Get(context.Background(), "k2")))
	require.Equal(t, 1, tests.NoError[int](t)(c.Get(context.Background(), "k2")))
	require.Equal(t, 0, tests.NoError[int](t)(c.Get(context.Background(), "k1")))

	time.Sleep(125 * time.Millisecond)
	require.Equal(t, 2, tests.NoError[int](t)(c.Get(context.Background(), "k2")))
	require.Equal(t, 3, tests.NoError[int](t)(c.Get(context.Background(), "k1")))
}
