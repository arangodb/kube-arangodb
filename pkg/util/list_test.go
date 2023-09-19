//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
