//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_PlanLocals(t *testing.T) {
	var l PlanLocals

	var key PlanLocalKey = "test"
	v1, v2 := "v1", "v2"

	t.Run("Get on nil", func(t *testing.T) {
		v, ok := l.Get(key)

		require.Equal(t, "", v)
		require.False(t, ok)
	})

	t.Run("Remove on nil", func(t *testing.T) {
		ok := l.Remove(key)

		require.False(t, ok)
	})

	t.Run("Add", func(t *testing.T) {
		ok := l.Add(key, v1, false)

		require.True(t, ok)

		v, ok := l.Get(key)

		require.True(t, ok)
		require.Equal(t, v1, v)
	})

	t.Run("Update", func(t *testing.T) {
		ok := l.Add(key, v2, false)

		require.False(t, ok)

		v, ok := l.Get(key)

		require.True(t, ok)
		require.Equal(t, v1, v)
	})

	t.Run("Update - override", func(t *testing.T) {
		ok := l.Add(key, v2, true)

		require.True(t, ok)

		v, ok := l.Get(key)

		require.True(t, ok)
		require.Equal(t, v2, v)
	})

	t.Run("Remove", func(t *testing.T) {
		ok := l.Remove(key)

		require.True(t, ok)
	})

	t.Run("Remove missing", func(t *testing.T) {
		ok := l.Remove(key)

		require.False(t, ok)
	})
}

func Test_PlanLocals_Equal(t *testing.T) {
	cmp := func(name string, a, b PlanLocals, expected bool) {
		t.Run(name, func(t *testing.T) {
			require.True(t, a.Equal(a))
			require.True(t, b.Equal(b))
			if expected {
				require.True(t, a.Equal(b))
				require.True(t, b.Equal(a))
			} else {
				require.False(t, a.Equal(b))
				require.False(t, b.Equal(a))
			}
		})
	}

	cmp("Nil", nil, nil, true)

	cmp("Nil & empty", nil, PlanLocals{}, true)

	cmp("Empty", PlanLocals{}, PlanLocals{}, true)

	cmp("Same keys & values", PlanLocals{
		"key1": "v1",
	}, PlanLocals{
		"key1": "v1",
	}, true)

	cmp("Diff keys", PlanLocals{
		"key2": "v1",
	}, PlanLocals{
		"key1": "v1",
	}, false)

	cmp("Same keys & diff values", PlanLocals{
		"key1": "v1",
	}, PlanLocals{
		"key1": "v2",
	}, false)

	cmp("Same multi keys & values", PlanLocals{
		"key1": "v1",
		"ket2": "v2",
	}, PlanLocals{
		"key1": "v1",
		"ket2": "v2",
	}, true)

	cmp("Same multi keys & values - reorder", PlanLocals{
		"key1": "v1",
		"ket2": "v2",
	}, PlanLocals{
		"ket2": "v2",
		"key1": "v1",
	}, true)
}
