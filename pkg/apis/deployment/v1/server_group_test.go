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

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerGroupAsRole(t *testing.T) {
	assert.Equal(t, "single", ServerGroupSingle.AsRole())
	assert.Equal(t, "agent", ServerGroupAgents.AsRole())
	assert.Equal(t, "dbserver", ServerGroupDBServers.AsRole())
	assert.Equal(t, "coordinator", ServerGroupCoordinators.AsRole())
	assert.Equal(t, "syncmaster", ServerGroupSyncMasters.AsRole())
	assert.Equal(t, "syncworker", ServerGroupSyncWorkers.AsRole())
}

func TestServerGroupAsRoleAbbreviated(t *testing.T) {
	assert.Equal(t, "sngl", ServerGroupSingle.AsRoleAbbreviated())
	assert.Equal(t, "agnt", ServerGroupAgents.AsRoleAbbreviated())
	assert.Equal(t, "prmr", ServerGroupDBServers.AsRoleAbbreviated())
	assert.Equal(t, "crdn", ServerGroupCoordinators.AsRoleAbbreviated())
	assert.Equal(t, "syma", ServerGroupSyncMasters.AsRoleAbbreviated())
	assert.Equal(t, "sywo", ServerGroupSyncWorkers.AsRoleAbbreviated())
}

func TestServerGroupIsArangod(t *testing.T) {
	assert.True(t, ServerGroupSingle.IsArangod())
	assert.True(t, ServerGroupAgents.IsArangod())
	assert.True(t, ServerGroupDBServers.IsArangod())
	assert.True(t, ServerGroupCoordinators.IsArangod())
	assert.False(t, ServerGroupSyncMasters.IsArangod())
	assert.False(t, ServerGroupSyncWorkers.IsArangod())
}

func TestServerGroupIsArangosync(t *testing.T) {
	assert.False(t, ServerGroupSingle.IsArangosync())
	assert.False(t, ServerGroupAgents.IsArangosync())
	assert.False(t, ServerGroupDBServers.IsArangosync())
	assert.False(t, ServerGroupCoordinators.IsArangosync())
	assert.True(t, ServerGroupSyncMasters.IsArangosync())
	assert.True(t, ServerGroupSyncWorkers.IsArangosync())
}
