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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func runCountWithMeasure(t *testing.T, c int, name string, f func(t *testing.T)) {
	t.Run(name, func(t *testing.T) {
		n := time.Now()
		defer func() {
			s := time.Since(n)
			t.Logf("Elapsed: %s - %s per item", s.String(), s/time.Duration(c))
		}()

		for i := 0; i < c; i++ {
			runWithMeasure(t, fmt.Sprintf("R%03d", i), f)
		}
	})
}

func runWithMeasure(t *testing.T, name string, f func(t *testing.T)) {
	t.Run(name, func(t *testing.T) {
		n := time.Now()
		defer func() {
			t.Logf("Elapsed: %s", time.Since(n).String())
		}()

		f(t)
	})
}

func perfWithSize(t *testing.T, dbs, collections, shards, rf, servers int) {
	t.Run(fmt.Sprintf("%d/%d/%d/%d/%d", dbs, collections, shards, rf, servers), func(t *testing.T) {
		var s State

		runWithMeasure(t, "Generate", func(t *testing.T) {
			s = generateDatabases(t, dbs, collections, shards, rf, servers)
		})

		var count int

		runWithMeasure(t, "CountShards", func(t *testing.T) {
			count = s.CountShards()
			t.Logf("Shard: %d", count)
		})

		runWithMeasure(t, "GetDBS", func(t *testing.T) {
			t.Logf("Servers: %d", len(s.PlanServers()))
		})

		runWithMeasure(t, "Restartable", func(t *testing.T) {
			for id := 0; id < servers; id++ {
				name := fmt.Sprintf("server-%d", id)
				runWithMeasure(t, name, func(t *testing.T) {
					require.Len(t, GetDBServerBlockingRestartShards(s, Server(name)), 0)
				})
			}
		})

		runWithMeasure(t, "NotInSync", func(t *testing.T) {
			for id := 0; id < servers; id++ {
				name := fmt.Sprintf("server-%d", id)
				runWithMeasure(t, name, func(t *testing.T) {
					require.Len(t, GetDBServerShardsNotInSync(s, Server(name)), 0)
				})
			}
		})

		runWithMeasure(t, "GlobalNotInSync", func(t *testing.T) {
			require.Len(t, GetDBServerShardsNotInSync(s, "*"), 0)
		})

		runWithMeasure(t, "All", func(t *testing.T) {
			require.Len(t, s.Filter(func(s State, db, col, shard string) bool {
				return true
			}), count)
		})
	})
}

func Test_Perf_Calc(t *testing.T) {
	perfWithSize(t, 1, 1, 1, 1, 1)
	perfWithSize(t, 1, 32, 32, 2, 3)
	perfWithSize(t, 32, 32, 32, 3, 32)
	perfWithSize(t, 128, 32, 32, 3, 32)
}

func generateDatabases(t *testing.T, dbs, collections, shards, rf, servers int) State {
	gens := make([]Generator, dbs)

	for id := 0; id < dbs; id++ {
		gens[id] = generateCollections(t, NewDatabaseRandomGenerator(), collections, shards, rf, servers).Add()
	}

	return GenerateState(t, gens...)
}

func generateCollections(t *testing.T, db DatabaseGeneratorInterface, collections, shards, rf, servers int) DatabaseGeneratorInterface {
	d := db

	for id := 0; id < collections; id++ {
		d = generateShards(t, d.RandomCollection(), shards, rf, servers).Add()
	}

	return d
}

func generateShards(t *testing.T, col CollectionGeneratorInterface, shards, rf, servers int) CollectionGeneratorInterface {
	c := col

	for id := 0; id < shards; id++ {
		l := getServersSublist(t, rf, servers)
		c = c.WithShard().WithPlan(l...).WithCurrent(l...).Add()
	}

	return c
}

func getServersSublist(t *testing.T, rf, servers int) Servers {
	require.NotEqual(t, 0, rf)
	if rf > servers {
		require.Fail(t, "Server count is smaller than rf")
	}

	return generateServersSublist(servers)[0:rf]
}

func generateServersSublist(servers int) Servers {
	s := make(Servers, servers)

	for id := range s {
		s[id] = Server(fmt.Sprintf("server-%d", id))
	}

	rand.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})

	return s
}
