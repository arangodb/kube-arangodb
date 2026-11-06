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

package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func name(n string) operation.ItemName {
	return operation.ItemName{
		Namespace: "default",
		Name:      n,
	}
}

func Test_Cache(t *testing.T) {
	t.Run("Missing entry", func(t *testing.T) {
		c := NewObjectCache[int](time.Minute)

		_, ok := c.Get(name("a"))
		require.False(t, ok)
	})

	t.Run("Set and Get", func(t *testing.T) {
		c := NewObjectCache[int](time.Minute)

		c.Set(name("a"), 7)

		v, ok := c.Get(name("a"))
		require.True(t, ok)
		require.Equal(t, 7, v)
	})

	t.Run("Set replaces", func(t *testing.T) {
		c := NewObjectCache[int](time.Minute)

		c.Set(name("a"), 1)
		c.Set(name("a"), 2)

		v, ok := c.Get(name("a"))
		require.True(t, ok)
		require.Equal(t, 2, v)
	})

	t.Run("Separate names", func(t *testing.T) {
		c := NewObjectCache[int](time.Minute)

		c.Set(name("a"), 1)
		c.Set(name("b"), 2)

		v, ok := c.Get(name("b"))
		require.True(t, ok)
		require.Equal(t, 2, v)
	})

	t.Run("Namespace distinguishes", func(t *testing.T) {
		c := NewObjectCache[int](time.Minute)

		c.Set(operation.ItemName{Namespace: "a", Name: "x"}, 1)
		c.Set(operation.ItemName{Namespace: "b", Name: "x"}, 2)

		v, ok := c.Get(operation.ItemName{Namespace: "a", Name: "x"})
		require.True(t, ok)
		require.Equal(t, 1, v)
	})

	t.Run("Entry expires after TTL", func(t *testing.T) {
		c := NewObjectCache[int](time.Millisecond)

		c.Set(name("a"), 1)
		time.Sleep(5 * time.Millisecond)

		_, ok := c.Get(name("a"))
		require.False(t, ok)
	})

	t.Run("Zero TTL never expires", func(t *testing.T) {
		c := NewObjectCache[int](0)

		c.Set(name("a"), 1)

		v, ok := c.Get(name("a"))
		require.True(t, ok)
		require.Equal(t, 1, v)
	})

	t.Run("Invalidate", func(t *testing.T) {
		c := NewObjectCache[int](time.Minute)

		c.Set(name("a"), 1)
		c.Invalidate(name("a"))

		_, ok := c.Get(name("a"))
		require.False(t, ok)
	})

	t.Run("List returns all entries", func(t *testing.T) {
		c := NewObjectCache[int](time.Minute)

		c.Set(name("a"), 1)
		c.Set(name("b"), 2)
		c.Set(operation.ItemName{Namespace: "other", Name: "a"}, 3)

		require.ElementsMatch(t, []int{1, 2, 3}, c.List())
	})

	t.Run("List is empty", func(t *testing.T) {
		c := NewObjectCache[int](time.Minute)

		require.Empty(t, c.List())
	})

	t.Run("List drops expired entries", func(t *testing.T) {
		c := NewObjectCache[int](time.Millisecond)

		c.Set(name("a"), 1)
		time.Sleep(5 * time.Millisecond)

		require.Empty(t, c.List())
	})

	t.Run("Clear", func(t *testing.T) {
		c := NewObjectCache[int](time.Minute)

		c.Set(name("a"), 1)
		c.Set(name("b"), 2)
		c.Clear()

		_, ok := c.Get(name("a"))
		require.False(t, ok)
		_, ok = c.Get(name("b"))
		require.False(t, ok)
	})
}
