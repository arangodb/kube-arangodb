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

package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func testDictDefaultValue[T any](t *testing.T, expected T) {
	t.Run(reflect.TypeOf(expected).String(), func(t *testing.T) {
		m := map[string]T{}

		ev, ok := m["missing"]
		require.False(t, ok)

		require.Equal(t, expected, ev)

		evs := m["missing"]

		require.Equal(t, expected, evs)
	})
}

func Test_Dict_Types(t *testing.T) {
	testDictDefaultValue[string](t, "")
	testDictDefaultValue[int](t, 0)
	testDictDefaultValue[*string](t, nil)
	testDictDefaultValue[*int](t, nil)
}
