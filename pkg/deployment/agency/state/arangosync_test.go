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

func Test_Sync_Target(t *testing.T) {
	var s DumpState
	require.NoError(t, json.Unmarshal(syncSource, &s))

	require.NotNil(t, s.Agency.ArangoDB.ArangoSync.ArangoSync)
	require.True(t, s.Agency.ArangoDB.ArangoSync.ArangoSync.State.Outgoing.Targets.Exists())
	require.True(t, s.Agency.ArangoDB.ArangoSync.IsSyncInProgress())
}

func Test_Sync_Source(t *testing.T) {
	var s DumpState
	require.NoError(t, json.Unmarshal(syncTarget, &s))

	require.NotNil(t, s.Agency.ArangoDB.ArangoSync.ArangoSync)
	require.NotNil(t, s.Agency.ArangoDB.ArangoSync.ArangoSync.State.Incoming.State)
	require.Equal(t, "running", *s.Agency.ArangoDB.ArangoSync.ArangoSync.State.Incoming.State)
	require.True(t, s.Agency.ArangoDB.ArangoSync.IsSyncInProgress())
}
