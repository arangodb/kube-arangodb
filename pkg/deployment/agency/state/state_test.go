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
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/agency_dump.3.6.json
var agencyDump36 []byte

//go:embed testdata/agency_dump.3.7.json
var agencyDump37 []byte

//go:embed testdata/agency_dump.3.8.json
var agencyDump38 []byte

//go:embed testdata/agency_dump.3.9.json
var agencyDump39 []byte

//go:embed testdata/agency_dump.3.9.satellite.json
var agencyDump39Satellite []byte

//go:embed testdata/agency_dump.3.9.hotbackup.json
var agencyDump39HotBackup []byte

//go:embed testdata/agency_dump.3.9.jobs.json
var agencyDump39Jobs []byte

//go:embed testdata/longdata.json
var longData []byte

//go:embed testdata/sync.source.json
var syncSource []byte

//go:embed testdata/sync.target.json
var syncTarget []byte

var (
	data = map[string][]byte{
		"3.6":           agencyDump36,
		"3.7":           agencyDump37,
		"3.8":           agencyDump38,
		"3.9":           agencyDump39,
		"3.9-satellite": agencyDump39Satellite,
		"3.9-hotbackup": agencyDump39HotBackup,
	}
)

func Test_Unmarshal_MultiVersion(t *testing.T) {
	for v, data := range data {
		t.Run(v, func(t *testing.T) {
			var s DumpState
			require.NoError(t, json.Unmarshal(data, &s))

			t.Run("Ensure Names", func(t *testing.T) {
				for _, collections := range s.Agency.Arango.Plan.Collections {
					for _, collection := range collections {
						require.NotNil(t, collection.Name)
					}
				}
			})

			t.Run("Ensure distributeShardsLike", func(t *testing.T) {
				collections, ok := s.Agency.Arango.Plan.Collections["_system"]
				require.True(t, ok)

				for _, collection := range collections {
					name := *collection.Name
					t.Run(name, func(t *testing.T) {
						if name == "_users" {
							require.Nil(t, collection.DistributeShardsLike)
							return
						}
						require.NotNil(t, collection.DistributeShardsLike)

						n, ok := collections[*collection.DistributeShardsLike]
						require.True(t, ok)

						require.NotNil(t, n.Name)
						require.Nil(t, n.DistributeShardsLike)
						require.Equal(t, "_users", *n.Name)
					})
				}
			})
		})
	}
}

func Test_Unmarshal_Jobs(t *testing.T) {
	var s DumpState
	require.NoError(t, json.Unmarshal(agencyDump39Jobs, &s))

	require.Len(t, s.Agency.Arango.Target.JobToDo, 2)
	require.Len(t, s.Agency.Arango.Target.JobFailed, 3)
	require.Len(t, s.Agency.Arango.Target.JobPending, 1)
	require.Len(t, s.Agency.Arango.Target.JobFinished, 4)

	t.Run("Check GetJob", func(t *testing.T) {
		t.Run("Unknown", func(t *testing.T) {
			j, s := s.Agency.Arango.Target.GetJob("955400")
			require.Equal(t, JobPhaseUnknown, s)
			require.Equal(t, "", j.Type)
		})
		t.Run("Failed", func(t *testing.T) {
			j, s := s.Agency.Arango.Target.GetJob("955410")
			require.Equal(t, JobPhaseFailed, s)
			require.Equal(t, "resignLeadership", j.Type)
		})
		t.Run("ToDo", func(t *testing.T) {
			j, s := s.Agency.Arango.Target.GetJob("955430")
			require.Equal(t, JobPhaseToDo, s)
			require.Equal(t, "resignLeadership", j.Type)
		})
		t.Run("Pending", func(t *testing.T) {
			j, s := s.Agency.Arango.Target.GetJob("955420")
			require.Equal(t, JobPhasePending, s)
			require.Equal(t, "resignLeadership", j.Type)
		})
		t.Run("Finished", func(t *testing.T) {
			j, s := s.Agency.Arango.Target.GetJob("955440")
			require.Equal(t, JobPhaseFinished, s)
			require.Equal(t, "resignLeadership", j.Type)
		})
	})
}

func Test_Unmarshal_LongData(t *testing.T) {
	var s []Root

	require.NoError(t, json.Unmarshal(longData, &s))

	t.Logf("%+v", s)
}

func Test_IsDBServerInSync(t *testing.T) {
	type testCase struct {
		generator Generator
		inSync    []string
		notInSync []string
	}
	newDBWithCol := func(writeConcern int) CollectionGeneratorInterface {
		return NewDatabaseRandomGenerator().RandomCollection().WithWriteConcern(writeConcern)
	}
	tcs := map[string]testCase{
		"in Plan, in Current, WC = 2": {
			generator: newDBWithCol(2).
				WithShard().WithPlan("A", "B").WithCurrent("B").Add().
				WithShard().WithPlan("A", "B", "C").WithCurrent("A", "B", "C").Add().Add().Add(),
			inSync:    []string{"A", "C", "D"},
			notInSync: []string{"B"},
		},
		"in Plan, in Current, WC = 3, broken": {
			generator: newDBWithCol(3).
				WithShard().WithPlan("A", "B", "C").WithCurrent("A", "B").Add().
				WithShard().WithPlan("A", "B", "C").WithCurrent("A", "B", "C").Add().Add().Add(),
			inSync:    []string{"C"},
			notInSync: []string{"A", "B"},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			s := GenerateState(t, tc.generator)
			t.Run("InSync", func(t *testing.T) {
				for _, server := range tc.inSync {
					t.Run(server, func(t *testing.T) {
						require.Len(t, GetDBServerBlockingRestartShards(s, Server(server)), 0, "server %s should be in sync", server)
					})
				}
			})
			t.Run("NotInSync", func(t *testing.T) {
				for _, server := range tc.notInSync {
					t.Run(server, func(t *testing.T) {
						require.NotEqual(t, GetDBServerBlockingRestartShards(s, Server(server)), 0, "server %s should not be in sync", server)
					})
				}
			})
		})
	}
}

func Test_IsDBServerReadyToRestart(t *testing.T) {
	type testCase struct {
		generator Generator
		ready     []string
		notReady  []string
	}
	newDBWithCol := func(writeConcern int) CollectionGeneratorInterface {
		return NewDatabaseRandomGenerator().RandomCollection().WithWriteConcern(writeConcern)
	}
	tcs := map[string]testCase{
		"not in Plan, in Current": {
			generator: newDBWithCol(1).WithShard().WithPlan("A").WithCurrent("B").Add().Add().Add(),
			ready:     []string{"B", "A"},
		},
		"not in Plan, not in Current": {
			generator: newDBWithCol(1).WithShard().WithPlan("A").WithCurrent("A").Add().Add().Add(),
			ready:     []string{"C", "A"},
		},
		"in Plan and WC == RF": {
			generator: newDBWithCol(1).WithShard().WithPlan("A").WithCurrent("A").Add().Add().Add(),
			ready:     []string{"A"},
		},
		"in Plan, the only in Current": {
			generator: newDBWithCol(1).WithShard().WithPlan("A", "B").WithCurrent("A").Add().Add().Add(),
			ready:     []string{"B"},
			notReady:  []string{"A"},
		},
		"in Plan, missing in Current": {
			generator: newDBWithCol(1).WithShard().WithPlan("A", "B").WithCurrent("B").Add().Add().Add(),
			ready:     []string{"A"},
			notReady:  []string{"B"},
		},
		"in Plan, missing in Current but broken WC": {
			generator: newDBWithCol(2).WithShard().WithPlan("A", "B", "C").WithCurrent("B").Add().Add().Add(),
			notReady:  []string{"A", "B", "C"},
		},
		"in Plan, missing in Current with fine WC": {
			generator: newDBWithCol(2).WithShard().WithPlan("A", "B", "C").WithCurrent("B", "C").Add().Add().Add(),
			ready:     []string{"A"},
			notReady:  []string{"B", "C"},
		},
		"in Plan, missing in Current with low WC": {
			generator: newDBWithCol(1).WithShard().WithPlan("A", "B", "C").WithCurrent("B", "A").Add().Add().Add(),
			ready:     []string{"A", "B", "C"},
		},
		"in Plan, in Current but broken WC": {
			generator: newDBWithCol(2).WithShard().WithPlan("A", "B", "C").WithCurrent("B", "A").Add().Add().Add(),
			notReady:  []string{"A", "B"},
			ready:     []string{"C"},
		},
		"in Plan, all shards in sync": {
			generator: newDBWithCol(1).
				WithShard().WithPlan("A", "B", "C").WithCurrent("B", "A").Add().
				WithShard().WithPlan("A", "D").WithCurrent("D", "A").Add().
				Add().Add(),
			ready: []string{"A", "B", "C", "D"},
		},
		"all shards in sync but broken WC": {
			generator: newDBWithCol(2).
				WithShard().WithPlan("A", "B", "C").WithCurrent("B", "A").Add().
				WithShard().WithPlan("A", "D").WithCurrent("D", "A").Add().
				Add().Add(),
			notReady: []string{"A", "B"},
			ready:    []string{"C", "D"},
		},
		"some shards not fully synced": {
			generator: newDBWithCol(1).
				WithShard().WithPlan("A", "B", "C").WithCurrent("A", "B", "C").Add().
				WithShard().WithPlan("A", "B", "C").WithCurrent("C").Add().
				Add().Add(),
			ready:    []string{"A", "B"},
			notReady: []string{"C"},
		},
		"some shards not fully synced and broken WC": {
			generator: newDBWithCol(2).
				WithShard().WithPlan("A", "B", "C").WithCurrent("A", "B", "C").Add().
				WithShard().WithPlan("A", "B", "C").WithCurrent("C").Add().
				Add().Add(),
			notReady: []string{"A", "B", "C"},
		},
		"only one is able to restart": {
			generator: newDBWithCol(6).
				WithShard().WithPlan("A", "B", "C", "D", "E", "F").WithCurrent("A", "B", "C", "E", "F").Add().
				Add().Add(),
			ready:    []string{"D"},
			notReady: []string{"A", "B", "C", "E", "F"},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			s := GenerateState(t, tc.generator)
			for _, server := range tc.ready {
				require.Len(t, s.Filter(FilterDBServerShardRestart(Server(server))), 0, "server %s should be in sync", server)
			}
			for _, server := range tc.notReady {
				require.NotEqual(t, len(s.Filter(FilterDBServerShardRestart(Server(server)))), 0, "server %s should not be in sync", server)
			}
		})
	}
}
