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

package helm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testValuesMergeMethodMerge(t *testing.T, name string, a, b any, m ValuesMergeMethod, check func(t *testing.T, v Values, err error)) {
	t.Run(name, func(t *testing.T) {
		av, err := NewValues(a)
		require.NoError(t, err)

		bv, err := NewValues(b)
		require.NoError(t, err)

		v, err := m.Merge(av, bv)
		if check != nil {
			check(t, v, err)
		} else {
			require.NoError(t, err)
		}
	})
}

func Test_ValuesMergeMethod_Merge(t *testing.T) {
	testValuesMergeMethodMerge(t, "Nils", nil, nil, 0, func(t *testing.T, v Values, err error) {
		require.NoError(t, err)
		require.EqualValues(t, "{}", v.String())
	})
	testValuesMergeMethodMerge(t, "Second Nil", map[string]interface{}{
		"A": 1,
	}, nil, 0, func(t *testing.T, v Values, err error) {
		require.NoError(t, err)
		require.EqualValues(t, `{"A":1}`, v.String())
	})
	testValuesMergeMethodMerge(t, "First Nil", nil, map[string]interface{}{
		"B": 1,
	}, 0, func(t *testing.T, v Values, err error) {
		require.NoError(t, err)
		require.EqualValues(t, `{"B":1}`, v.String())
	})
	testValuesMergeMethodMerge(t, "Merge", map[string]interface{}{
		"A": 1,
	}, map[string]interface{}{
		"B": 1,
	}, 0, func(t *testing.T, v Values, err error) {
		require.NoError(t, err)
		require.EqualValues(t, `{"A":1,"B":1}`, v.String())
	})
	testValuesMergeMethodMerge(t, "Override", map[string]interface{}{
		"A": 1,
		"S": 0,
	}, map[string]interface{}{
		"B": 1,
		"S": 2,
	}, 0, func(t *testing.T, v Values, err error) {
		require.NoError(t, err)
		require.EqualValues(t, `{"A":1,"B":1,"S":2}`, v.String())
	})
	testValuesMergeMethodMerge(t, "Map - Default", map[string]interface{}{
		"M": map[string]interface{}{
			"A": 1,
			"S": 0,
		},
	}, map[string]interface{}{
		"M": map[string]interface{}{
			"B": 1,
			"S": 2,
		},
	}, 0, func(t *testing.T, v Values, err error) {
		require.NoError(t, err)
		require.EqualValues(t, `{"M":{"B":1,"S":2}}`, v.String())
	})
	testValuesMergeMethodMerge(t, "Map - Merge", map[string]interface{}{
		"M": map[string]interface{}{
			"A": 1,
			"S": 0,
		},
	}, map[string]interface{}{
		"M": map[string]interface{}{
			"B": 1,
			"S": 2,
		},
	}, MergeMaps, func(t *testing.T, v Values, err error) {
		require.NoError(t, err)
		require.EqualValues(t, `{"M":{"A":1,"B":1,"S":2}}`, v.String())
	})
	testValuesMergeMethodMerge(t, "SubMap - Default", map[string]interface{}{
		"M": map[string]interface{}{
			"M": map[string]interface{}{
				"A": 1,
				"S": 0,
			},
		},
	}, map[string]interface{}{
		"M": map[string]interface{}{
			"M": map[string]interface{}{
				"B": 1,
				"S": 2,
			},
		},
	}, 0, func(t *testing.T, v Values, err error) {
		require.NoError(t, err)
		require.EqualValues(t, `{"M":{"M":{"B":1,"S":2}}}`, v.String())
	})
	testValuesMergeMethodMerge(t, "SubMap - Merge", map[string]interface{}{
		"M": map[string]interface{}{
			"M": map[string]interface{}{
				"A": 1,
				"S": 0,
			},
		},
	}, map[string]interface{}{
		"M": map[string]interface{}{
			"M": map[string]interface{}{
				"B": 1,
				"S": 2,
			},
		},
	}, MergeMaps, func(t *testing.T, v Values, err error) {
		require.NoError(t, err)
		require.EqualValues(t, `{"M":{"M":{"A":1,"B":1,"S":2}}}`, v.String())
	})
}
