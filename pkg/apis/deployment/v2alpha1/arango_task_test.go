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
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ArangoTask_Details(t *testing.T) {
	arangoTaskDetails(t, int(5))
	arangoTaskDetails(t, "test")
	arangoTaskDetails(t, []interface{}{"data", "two"})
	arangoTaskDetails(t, map[string]interface{}{
		"data": "exp",
	})
}

func arangoTaskDetails(t *testing.T, obj interface{}) {
	arangoTaskDetailsExp(t, obj, obj)
}

func arangoTaskDetailsExp(t *testing.T, obj, exp interface{}) {
	t.Run(reflect.TypeOf(obj).String(), func(t *testing.T) {
		var d ArangoTaskDetails

		require.NoError(t, d.Set(obj))

		b, err := json.Marshal(d)
		require.NoError(t, err)

		var n ArangoTaskDetails
		require.NoError(t, json.Unmarshal(b, &n))

		require.NoError(t, n.Get(&exp))

		require.EqualValues(t, obj, exp)
	})
}
