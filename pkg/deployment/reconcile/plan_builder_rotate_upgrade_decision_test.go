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

package reconcile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func buildUpdateUpgradeDecisionMap(elems ...updateUpgradeDecision) updateUpgradeDecisionMap {
	m := updateUpgradeDecisionMap{}
	for id, e := range elems {
		m[updateUpgradeDecisionItem{
			ID:    fmt.Sprintf("Server%05d", id),
			Group: api.ServerGroupDBServers,
		}] = e
	}

	return m
}

func Test_updateUpgradeDecisionMap_GetFromToVersion(t *testing.T) {
	t.Run("With Empty", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap().GetFromToVersion()
		require.EqualValues(t, "", from)
		require.EqualValues(t, "", to)
	})

	t.Run("With NonUpgrade", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap(updateUpgradeDecision{
			upgrade: false,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: false,
				FromVersion:   "3.12.3",
				ToVersion:     "3.12.4",
			},
		}).GetFromToVersion()
		require.EqualValues(t, "", from)
		require.EqualValues(t, "", to)
	})

	t.Run("With Upgrade", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap(updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				FromVersion:   "3.12.3",
				ToVersion:     "3.12.4",
			},
		}).GetFromToVersion()
		require.EqualValues(t, "3.12.3", from)
		require.EqualValues(t, "3.12.4", to)
	})

	t.Run("With Upgrade of single member", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap(updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				FromVersion:   "3.12.3",
				ToVersion:     "3.12.4",
			},
		}, updateUpgradeDecision{
			upgrade: false,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: false,
				FromVersion:   "3.12.5",
				ToVersion:     "3.12.5",
			},
		}).GetFromToVersion()
		require.EqualValues(t, "3.12.3", from)
		require.EqualValues(t, "3.12.4", to)
	})

	t.Run("With Upgrade of multi member", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap(updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				FromVersion:   "3.12.3",
				ToVersion:     "3.12.4",
			},
		}, updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				FromVersion:   "3.12.2",
				ToVersion:     "3.12.3",
			},
		}, updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				FromVersion:   "3.12.2",
				ToVersion:     "3.12.5",
			},
		}).GetFromToVersion()
		require.EqualValues(t, "3.12.2", from)
		require.EqualValues(t, "3.12.5", to)
	})
}
