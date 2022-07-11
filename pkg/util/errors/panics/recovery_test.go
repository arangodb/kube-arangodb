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

package panics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func Test_Panic(t *testing.T) {
	tests.WithLogScanner(t, "Panic", func(t *testing.T, s tests.LogScanner) {
		g := s.Factory().RegisterAndGetLogger("panic", logging.Error)
		err := RecoverWithSectionL(g, "foo", func() error {
			panic(3)
		})

		d, ok := IsPanicError(err)

		require.True(t, ok)

		var data map[string]interface{}

		require.True(t, s.GetData(t, 100*time.Millisecond, &data))
		t.Run("stack", func(t *testing.T) {
			require.Contains(t, data, "stack")

			stack, ok := data["stack"].([]interface{})
			require.True(t, ok)

			require.Len(t, stack, len(d.Stack()))

			for id := range stack {
				s, ok := stack[id].(string)
				require.True(t, ok)

				require.Equal(t, d.Stack()[id].String(), s)
			}
		})

		t.Run("section", func(t *testing.T) {
			require.Contains(t, data, "section")

			s, ok := data["section"].(string)
			require.True(t, ok)

			require.Equal(t, "foo", s)
		})

		t.Run("value", func(t *testing.T) {
			require.EqualValues(t, 3, d.PanicCause())
		})
	})
}
