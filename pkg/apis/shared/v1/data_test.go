//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_Data(t *testing.T) {
	var d Data = make([]byte, 1024)

	z, err := util.JSONRemarshal[Data, Data](d)
	require.NoError(t, err)
	require.EqualValues(t, d, z)
}

func testDataRepresentation[IN any](t *testing.T, v IN) {
	t.Run(reflect.TypeFor[IN]().String(), func(t *testing.T) {
		d, err := NewData[IN](v)
		require.NoError(t, err)

		z, err := json.Marshal(d)
		require.NoError(t, err)

		t.Logf("%s", string(z))

		var d2 Data

		require.NoError(t, json.Unmarshal(z, &d2))

		require.EqualValues(t, d, d2)
	})
}

func Test_Data_Representation(t *testing.T) {
	testDataRepresentation(t, "some_Test_data")
	testDataRepresentation(t, []string{"A", "B"})
	testDataRepresentation(t, 53)
	testDataRepresentation(t, Object{
		Name: "X",
	})
}
