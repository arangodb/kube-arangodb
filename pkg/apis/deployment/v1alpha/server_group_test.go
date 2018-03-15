//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package v1alpha

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerGroupAsRole(t *testing.T) {
	assert.Equal(t, "sngl", ServerGroupSingle.AsRole())
	assert.Equal(t, "agnt", ServerGroupAgents.AsRole())
	assert.Equal(t, "prmr", ServerGroupDBServers.AsRole())
	assert.Equal(t, "crdn", ServerGroupCoordinators.AsRole())
	assert.Equal(t, "syma", ServerGroupSyncMasters.AsRole())
	assert.Equal(t, "sywo", ServerGroupSyncWorkers.AsRole())
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
