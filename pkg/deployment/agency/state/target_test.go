//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package state

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Target_HotBackup(t *testing.T) {
	t.Run("Exists", func(t *testing.T) {
		var s DumpState
		require.NoError(t, json.Unmarshal(agencyDump39HotBackup, &s))

		require.True(t, s.Agency.Arango.Target.HotBackup.Create.Exists())

		t.Log(s.Agency.Arango.Target.HotBackup.Create.time.String())

		require.False(t, s.Agency.ArangoDB.ArangoSync.IsSyncInProgress())
	})
	t.Run("Does Not Exists", func(t *testing.T) {
		var s DumpState
		require.NoError(t, json.Unmarshal(agencyDump39Satellite, &s))

		require.False(t, s.Agency.Arango.Target.HotBackup.Create.Exists())

		require.False(t, s.Agency.ArangoDB.ArangoSync.IsSyncInProgress())
	})
}
