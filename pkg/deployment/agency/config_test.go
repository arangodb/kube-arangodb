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

package agency

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/config.json
var config []byte

func Test_Config_Unmarshal(t *testing.T) {
	var cfg Config

	require.NoError(t, json.Unmarshal(config, &cfg))

	require.Equal(t, "AGNT-fd0f4fc7-b60b-44bb-9f5e-5fc91f708f82", cfg.LeaderId)
	require.Equal(t, uint64(94), cfg.CommitIndex)
	require.Equal(t, "AGNT-fd0f4fc7-b60b-44bb-9f5e-5fc91f708f82", cfg.Configuration.ID)
}
