//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_List_Sort(t *testing.T) {
	type obj struct {
		creationDate time.Time
	}
	now := time.Now()

	l := List[*obj]{
		&obj{now},
		&obj{now.Add(time.Second)},
		&obj{now.Add(-time.Second)},
		&obj{now.Add(time.Hour)},
		&obj{now.Add(-time.Hour)},
	}
	expected := List[*obj]{
		&obj{now.Add(time.Hour)},
		&obj{now.Add(time.Second)},
		&obj{now},
		&obj{now.Add(-time.Second)},
		&obj{now.Add(-time.Hour)},
	}
	sorted := l.Sort(func(a *obj, b *obj) bool {
		return a.creationDate.After(b.creationDate)
	})
	require.EqualValues(t, expected, sorted)
}

func Test_MapList(t *testing.T) {
	type obj struct {
		name string
	}
	l := List[*obj]{
		&obj{"a"},
		&obj{"b"},
		&obj{"c"},
	}
	expected := List[string]{"a", "b", "c"}
	require.Equal(t, expected, MapList(l, func(o *obj) string {
		return o.name
	}))
}

func Test_AppendAfter(t *testing.T) {
	var elements []int

	elements = AppendAfter(elements, func(v int) bool {
		return false
	}, 1)
	require.Equal(t, []int{1}, elements)

	elements = AppendAfter(elements, func(v int) bool {
		return false
	}, 2)
	require.Equal(t, []int{1, 2}, elements)

	elements = AppendAfter(elements, func(v int) bool {
		return v == 1
	}, 3)
	require.Equal(t, []int{1, 3, 2}, elements)

	elements = AppendAfter(elements, func(v int) bool {
		return v == 2
	}, 4)
	require.Equal(t, []int{1, 3, 2, 4}, elements)
}

func Test_Batcher(t *testing.T) {
	v := make([]int, 17)
	for id := range v {
		v[id] = id
	}

	res := BatchList(16, v)
	require.Len(t, res, 2)

	require.Len(t, res[0], 16)
	require.Len(t, res[1], 1)

	require.Equal(t, v[0:16], res[0])
	require.Equal(t, v[16:], res[1])
}
