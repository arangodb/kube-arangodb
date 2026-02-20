//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/util"
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

func Test_updateUpgradeDecisionMap_GetFromTo(t *testing.T) {
	t.Run("With Empty", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap().GetFromTo()
		require.EqualValues(t, "", from)
		require.EqualValues(t, "", to)
	})

	t.Run("With NonUpgrade", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap(updateUpgradeDecision{
			upgrade: false,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: false,
				From:          api.ImageInfo{ArangoDBVersion: "3.12.3"},
				To:            api.ImageInfo{ArangoDBVersion: "3.12.4"},
			},
		}).GetFromTo()
		require.EqualValues(t, "", from)
		require.EqualValues(t, "", to)
	})

	t.Run("With Upgrade", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap(updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				From:          api.ImageInfo{ArangoDBVersion: "3.12.3"},
				To:            api.ImageInfo{ArangoDBVersion: "3.12.4"},
			},
		}).GetFromTo()
		require.EqualValues(t, "3.12.3", from)
		require.EqualValues(t, "3.12.4", to)
	})

	t.Run("With Upgrade of single member", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap(updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				From:          api.ImageInfo{ArangoDBVersion: "3.12.3"},
				To:            api.ImageInfo{ArangoDBVersion: "3.12.4"},
			},
		}, updateUpgradeDecision{
			upgrade: false,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: false,
				From:          api.ImageInfo{ArangoDBVersion: "3.12.5"},
				To:            api.ImageInfo{ArangoDBVersion: "3.12.5"},
			},
		}).GetFromTo()
		require.EqualValues(t, "3.12.3", from)
		require.EqualValues(t, "3.12.4", to)
	})

	t.Run("With Upgrade of multi member", func(t *testing.T) {
		from, to := buildUpdateUpgradeDecisionMap(updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				From:          api.ImageInfo{ArangoDBVersion: "3.12.3"},
				To:            api.ImageInfo{ArangoDBVersion: "3.12.4"},
			},
		}, updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				From:          api.ImageInfo{ArangoDBVersion: "3.12.2"},
				To:            api.ImageInfo{ArangoDBVersion: "3.12.3"},
			},
		}, updateUpgradeDecision{
			upgrade: true,
			upgradeDecision: upgradeDecision{
				UpgradeNeeded: true,
				From:          api.ImageInfo{ArangoDBVersion: "3.12.2"},
				To:            api.ImageInfo{ArangoDBVersion: "3.12.5"},
			},
		}).GetFromTo()
		require.EqualValues(t, "3.12.2", from)
		require.EqualValues(t, "3.12.5", to)
	})
}

func Test_CheckUpgradeRules(t *testing.T) {

	type testCase struct {
		Allowed  bool
		From, To api.ImageInfo
	}

	var testCases = []testCase{
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      false,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.11.1",
				Enterprise:      false,
			},
		},
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      false,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.11.1",
				Enterprise:      true,
			},
		},
		{
			Allowed: false,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.11.1",
				Enterprise:      false,
			},
		},
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      false,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.12.1",
				Enterprise:      false,
			},
		},
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      false,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.12.1",
				Enterprise:      true,
			},
		},
		{
			Allowed: false,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.12.1",
				Enterprise:      false,
			},
		},
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      false,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      true,
			},
		},
		{
			Allowed: false,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      false,
			},
		},
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "3.12.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "4.0.0",
				Enterprise:      true,
			},
		},
		{
			Allowed: false,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "4.0.0",
				Enterprise:      true,
			},
		},
		{
			Allowed: false,
			From: api.ImageInfo{
				ArangoDBVersion: "3.10.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.12.0",
				Enterprise:      true,
			},
		},
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "3.11.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "3.12.0",
				Enterprise:      true,
			},
		},
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "4.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "4.1",
				Enterprise:      true,
			},
		},
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "4.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "4.8",
				Enterprise:      true,
			},
		},
		{
			Allowed: true,
			From: api.ImageInfo{
				ArangoDBVersion: "4.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "5.7",
				Enterprise:      true,
			},
		},
		{
			Allowed: false,
			From: api.ImageInfo{
				ArangoDBVersion: "4.0",
				Enterprise:      true,
			},
			To: api.ImageInfo{
				ArangoDBVersion: "6.0",
				Enterprise:      true,
			},
		},
	}

	for _, v := range testCases {
		t.Run(fmt.Sprintf("%s (%s) -> %s (%s) : %s", v.From.ArangoDBVersion, util.BoolSwitch(v.From.Enterprise, "EE", "CE"), v.To.ArangoDBVersion, util.BoolSwitch(v.To.Enterprise, "EE", "CE"), util.BoolSwitch(v.Allowed, "Allowed", "Denied")), func(t *testing.T) {
			err := checkUpgradeRules(v.From, v.To)
			if v.Allowed {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
