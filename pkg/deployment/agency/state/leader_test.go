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
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LeaderCheck_DBServerIsLeader(t *testing.T) {
	t.Run("Empty collection", func(t *testing.T) {
		p := PlanCollections{}

		require.False(t, p.IsDBServerLeader("A"))
	})
	t.Run("Empty shards", func(t *testing.T) {
		p := PlanCollections{
			"a": {
				"a": {
					Shards: nil,
				},
			},
		}

		require.False(t, p.IsDBServerLeader("A"))
	})
	t.Run("Empty shard list", func(t *testing.T) {
		p := PlanCollections{
			"a": {
				"a": {
					Shards: Shards{
						"a": {},
					},
				},
			},
		}

		require.False(t, p.IsDBServerLeader("A"))
	})
	t.Run("Follower", func(t *testing.T) {
		p := PlanCollections{
			"a": {
				"a": {
					Shards: Shards{
						"a": {
							"B",
							"A",
						},
					},
				},
			},
		}

		require.False(t, p.IsDBServerLeader("A"))
	})
	t.Run("Leader", func(t *testing.T) {
		p := PlanCollections{
			"a": {
				"a": {
					Shards: Shards{
						"a": {
							"A",
							"B",
						},
					},
				},
			},
		}

		require.True(t, p.IsDBServerLeader("A"))
	})
}
