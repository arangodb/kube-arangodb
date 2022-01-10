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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func backoffParse(t *testing.T, in BackOff) BackOff {
	d, err := json.Marshal(in)
	require.NoError(t, err)

	var m BackOff

	require.NoError(t, json.Unmarshal(d, &m))

	return m
}

func Test_BackOff_Combine(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		a := BackOff{}
		b := BackOff{
			"a": meta.Now(),
		}

		r := a.Combine(b)

		require.Contains(t, r, BackOffKey("a"))

		r = backoffParse(t, r)

		require.Contains(t, r, BackOffKey("a"))
		require.Equal(t, b["a"].Unix(), r["a"].Unix())
	})

	t.Run("replace", func(t *testing.T) {
		a := BackOff{
			"a": meta.Time{Time: time.Now().Add(-1 * time.Hour)},
		}
		b := BackOff{
			"a": meta.Now(),
		}

		r := a.Combine(b)

		require.Contains(t, r, BackOffKey("a"))

		r = backoffParse(t, r)

		require.Contains(t, r, BackOffKey("a"))
		require.Equal(t, b["a"].Unix(), r["a"].Unix())
	})

	t.Run("delete", func(t *testing.T) {
		a := BackOff{
			"a": meta.Time{Time: time.Now().Add(-1 * time.Hour)},
		}
		b := BackOff{
			"a": meta.Time{},
		}

		r := a.Combine(b)

		require.Contains(t, r, BackOffKey("a"))

		r = backoffParse(t, r)

		require.NotContains(t, r, BackOffKey("a"))
	})
}
