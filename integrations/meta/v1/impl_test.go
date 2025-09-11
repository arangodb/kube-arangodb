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

var existing_json_data = `{"_key":"genai.all_project_names","_id":"_meta_store/genai.all_project_names","_rev":"_j8TwrTm--_","meta":{"updatedAt":"2025-07-07T15:13:09Z"},"object":{"Object":{"type_url":"type.googleapis.com/arangodb_platform_internal.metadata_store.GenAiProjectNames","value":"ChV0ZXN0X3Byb2plY3RfNmVhYWM3MjM="}}}`

func Test(t *testing.T) {
	var o Object
	err := json.Unmarshal([]byte(existing_json_data), &o)
	require.NoError(t, err)

	data, err := json.Marshal(&o)
	require.NoError(t, err)

	require.Equal(t, existing_json_data, string(data))
}
