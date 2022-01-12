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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Config_Unmarshal(t *testing.T) {
	data := `{
  "term": 0,
  "leaderId": "AGNT-fd0f4fc7-b60b-44bb-9f5e-5fc91f708f82",
  "commitIndex": 94,
  "lastCompactionAt": 0,
  "nextCompactionAfter": 500,
  "lastAcked": {
    "AGNT-fd0f4fc7-b60b-44bb-9f5e-5fc91f708f82": {
      "lastAckedTime": 0,
      "lastAckedIndex": 94
    }
  },
  "configuration": {
    "pool": {
      "AGNT-fd0f4fc7-b60b-44bb-9f5e-5fc91f708f82": "tcp://[::1]:4001"
    },
    "active": [
      "AGNT-fd0f4fc7-b60b-44bb-9f5e-5fc91f708f82"
    ],
    "id": "AGNT-fd0f4fc7-b60b-44bb-9f5e-5fc91f708f82",
    "agency size": 1,
    "pool size": 1,
    "endpoint": "tcp://[::1]:4001",
    "min ping": 1,
    "max ping": 5,
    "timeoutMult": 1,
    "supervision": true,
    "supervision frequency": 1,
    "compaction step size": 500,
    "compaction keep size": 50000,
    "supervision grace period": 10,
    "supervision ok threshold": 5,
    "version": 2,
    "startup": "origin"
  },
  "engine": "rocksdb",
  "version": "3.10.0-devel"
}`

	var cfg agencyConfig

	require.NoError(t, json.Unmarshal([]byte(data), &cfg))

	require.Equal(t, "AGNT-fd0f4fc7-b60b-44bb-9f5e-5fc91f708f82", cfg.LeaderId)
	require.Equal(t, uint64(94), cfg.CommitIndex)
	require.Equal(t, "AGNT-fd0f4fc7-b60b-44bb-9f5e-5fc91f708f82", cfg.Configuration.ID)
}
