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

package features

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func testIsUpgradeIndexOrderIssueEnabled(enabled bool, group api.ServerGroup, from, to driver.Version) bool {
	*upgradeIndexOrderIssue.EnabledPointer() = enabled

	return IsUpgradeIndexOrderIssueEnabled(group, from, to)
}

func Test_IsUpgradeIndexOrderIssueEnabled(t *testing.T) {
	require.True(t, testIsUpgradeIndexOrderIssueEnabled(true, api.ServerGroupDBServers, "3.12.2", "3.12.4"))
	require.False(t, testIsUpgradeIndexOrderIssueEnabled(false, api.ServerGroupDBServers, "3.12.2", "3.12.4"))
	require.False(t, testIsUpgradeIndexOrderIssueEnabled(true, api.ServerGroupCoordinators, "3.12.2", "3.12.4"))
	require.False(t, testIsUpgradeIndexOrderIssueEnabled(true, api.ServerGroupAgents, "3.12.2", "3.12.4"))
	require.False(t, testIsUpgradeIndexOrderIssueEnabled(true, api.ServerGroupSingle, "3.12.2", "3.12.4"))
	require.True(t, testIsUpgradeIndexOrderIssueEnabled(true, api.ServerGroupDBServers, "3.12.3", "3.12.4"))
	require.False(t, testIsUpgradeIndexOrderIssueEnabled(true, api.ServerGroupDBServers, "3.12.4", "3.12.4"))
	require.True(t, testIsUpgradeIndexOrderIssueEnabled(true, api.ServerGroupDBServers, "3.12.2", "3.12.55"))
}
