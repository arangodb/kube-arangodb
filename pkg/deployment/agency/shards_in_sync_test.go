//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ShardsInSync(t *testing.T) {
	s := State{
		Current: StateCurrent{
			Collections: map[string]StateCurrentDBCollections{
				"a": map[string]StateCurrentDBCollection{
					"a": map[string]StateCurrentDBShard{
						"s0001": {
							Servers: Servers{
								"A",
								"B",
							},
						},
					},
				},
			},
		},
	}

	t.Run("All in sync", func(t *testing.T) {
		require.True(t, s.IsShardInSync("a", "a", "s0001", Servers{"A", "B"}))
	})

	t.Run("Invalid order", func(t *testing.T) {
		require.False(t, s.IsShardInSync("a", "a", "s0001", Servers{"B", "A"}))
	})

	t.Run("Missing server", func(t *testing.T) {
		require.False(t, s.IsShardInSync("a", "a", "s0001", Servers{"A"}))
	})

	t.Run("Missing db", func(t *testing.T) {
		require.False(t, s.IsShardInSync("a1", "a", "s0001", Servers{"A", "B"}))
	})

	t.Run("Missing col", func(t *testing.T) {
		require.False(t, s.IsShardInSync("a", "a1", "s0001", Servers{"A", "B"}))
	})

	t.Run("Missing shard", func(t *testing.T) {
		require.False(t, s.IsShardInSync("a", "a", "s00011", Servers{"A", "B"}))
	})
}
