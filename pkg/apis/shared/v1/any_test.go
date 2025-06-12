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

package v1

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func testAnyRepresentation[IN any](t *testing.T, v IN) {
	t.Run(reflect.TypeFor[IN]().String(), func(t *testing.T) {
		d, err := NewAny[IN](v)
		require.NoError(t, err)

		z, err := json.Marshal(d)
		require.NoError(t, err)

		t.Logf("%s", string(z))

		var d2 Any

		require.NoError(t, json.Unmarshal(z, &d2))

		require.EqualValues(t, d, d2)
	})
}

func Test_Any_Representation(t *testing.T) {
	testAnyRepresentation(t, "some_Test_Any")
	testAnyRepresentation(t, []string{"A", "B"})
	testAnyRepresentation(t, 53)
	testAnyRepresentation(t, Object{
		Name: "X",
	})
}
