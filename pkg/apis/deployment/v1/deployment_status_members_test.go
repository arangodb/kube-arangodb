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

	"github.com/stretchr/testify/require"
)

func newMemberList() DeploymentStatusMembers {
	return DeploymentStatusMembers{
		Single:       MemberStatusList{{ID: ServerGroupSingle.AsRole()}},
		Agents:       MemberStatusList{{ID: ServerGroupAgents.AsRole()}},
		DBServers:    MemberStatusList{{ID: ServerGroupDBServers.AsRole()}},
		Coordinators: MemberStatusList{{ID: ServerGroupCoordinators.AsRole()}},
		SyncMasters:  MemberStatusList{{ID: ServerGroupSyncMasters.AsRole()}},
		SyncWorkers:  MemberStatusList{{ID: ServerGroupSyncWorkers.AsRole()}},
	}
}

func Test_StatusMemberList_EnsureDefaultExecutionOrder(t *testing.T) {
	statusMembers := newMemberList()
	order := AllServerGroups
	orderIndex := 0

	for _, e := range statusMembers.AsList() {
		require.True(t, orderIndex < len(order))
		require.Equal(t, order[orderIndex], e.Group)
		require.Equal(t, order[orderIndex].AsRole(), e.Member.ID)

		orderIndex += 1
	}
}

func Test_StatusMemberList_CustomExecutionOrder(t *testing.T) {
	statusMembers := newMemberList()
	order := []ServerGroup{
		ServerGroupDBServers,
	}
	orderIndex := 0

	for _, e := range statusMembers.AsListInGroups(order...) {
		require.True(t, orderIndex < len(order))
		require.Equal(t, order[orderIndex], e.Group)
		require.Equal(t, order[orderIndex].AsRole(), e.Member.ID)

		orderIndex += 1
	}
}
