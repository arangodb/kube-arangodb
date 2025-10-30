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
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Types(t *testing.T) {
	z := `{
  "meta": {
    "updatedAt": "2025-07-07T15:13:09Z"
  },
  "object": {
    "Object": {
      "type_url": "type.googleapis.com/arangodb_platform_internal.metadata_store.GenAiProjectNames",
      "value": "ChV0ZXN0X3Byb2plY3RfNmVhYWM3MjM="
    }
  }
}`

	var obj Object

	require.NoError(t, json.Unmarshal([]byte(z), &obj))

	n, err := json.Marshal(obj)
	require.NoError(t, err)

	require.JSONEq(t, z, string(n))
}
