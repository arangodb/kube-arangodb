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

package agency

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

var (
	data = map[string][]byte{
		"3.6": agencyDump36,
		"3.7": agencyDump37,
		"3.8": agencyDump38,
		"3.9": agencyDump39,
	}
)

func Test_Unmarshal_MultiVersion(t *testing.T) {
	for v, data := range data {
		t.Run(v, func(t *testing.T) {
			var s DumpState
			require.NoError(t, json.Unmarshal(data, &s))

			s.Agency.Arango.IterateOverCollections(func(db, col string, info *StatePlanCollection, shard string, plan ShardServers, current ShardServers) bool {
				require.EqualValues(t, 1, info.GetWriteConcern(0))
				return false
			})
		})
	}
}

func Test_Unmarshal_LongData(t *testing.T) {
	data := "[{\"arango\":{\"Supervision\":{},\"Current\":{\"Collections\":{\"_system\":{\"10011\":{\"s10022\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010022\",\"sparse\":false,\"type\":\"primary\",\"unique\":true},{\"deduplicate\":true,\"estimates\":true,\"fields\":[\"mount\"],\"id\":\"10029\",\"name\":\"idx_1718347303809449984\",\"objectId\":\"2010164\",\"sparse\":true,\"type\":\"hash\",\"unique\":true}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10005\":{\"s10016\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010038\",\"sparse\":false,\"type\":\"primary\",\"unique\":true},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"time\"],\"id\":\"10027\",\"name\":\"idx_1718347303741292544\",\"objectId\":\"2010144\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10012\":{\"s10023\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010032\",\"sparse\":false,\"type\":\"primary\",\"unique\":true}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10010\":{\"s10021\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010034\",\"sparse\":false,\"type\":\"primary\",\"unique\":true},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"queue\",\"status\",\"delayUntil\"],\"id\":\"10030\",\"name\":\"idx_1718347303839858688\",\"objectId\":\"2010174\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"status\",\"queue\",\"delayUntil\"],\"id\":\"10031\",\"name\":\"idx_1718347303866073088\",\"objectId\":\"2010186\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10004\":{\"s10015\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010036\",\"sparse\":false,\"type\":\"primary\",\"unique\":true},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"time\"],\"id\":\"10026\",\"name\":\"idx_1718347303708786688\",\"objectId\":\"2010134\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10003\":{\"s10014\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010028\",\"sparse\":false,\"type\":\"primary\",\"unique\":true}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10006\":{\"s10017\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010030\",\"sparse\":false,\"type\":\"primary\",\"unique\":true},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"time\"],\"id\":\"10028\",\"name\":\"idx_1718347303770652672\",\"objectId\":\"2010154\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10001\":{\"s10002\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010002\",\"sparse\":false,\"type\":\"primary\",\"unique\":true},{\"deduplicate\":true,\"estimates\":true,\"fields\":[\"user\"],\"id\":\"10025\",\"name\":\"idx_1718347303681523712\",\"objectId\":\"2010124\",\"sparse\":true,\"type\":\"hash\",\"unique\":true}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10007\":{\"s10018\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010027\",\"sparse\":false,\"type\":\"primary\",\"unique\":true}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10008\":{\"s10019\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010024\",\"sparse\":false,\"type\":\"primary\",\"unique\":true}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10009\":{\"s10020\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010040\",\"sparse\":false,\"type\":\"primary\",\"unique\":true}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}},\"10013\":{\"s10024\":{\"failoverCandidates\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"],\"errorNum\":0,\"errorMessage\":\"\",\"error\":false,\"indexes\":[{\"fields\":[\"_key\"],\"id\":\"0\",\"name\":\"primary\",\"objectId\":\"2010042\",\"sparse\":false,\"type\":\"primary\",\"unique\":true}],\"servers\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]}}}}},\"Plan\":{\"Collections\":{\"_system\":{\"10011\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10022\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_apps\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10011\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"mount\"],\"id\":\"10029\",\"inBackground\":false,\"name\":\"idx_1718347303809449984\",\"sparse\":true,\"type\":\"hash\",\"unique\":true}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10008\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10019\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_aqlfunctions\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10008\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10001\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"shardingStrategy\":\"hash\",\"shards\":{\"s10002\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"waitForSync\":false,\"schema\":null,\"shardKeys\":[\"_key\"],\"isDisjoint\":false,\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"user\"],\"id\":\"10025\",\"inBackground\":false,\"name\":\"idx_1718347303681523712\",\"sparse\":true,\"type\":\"hash\",\"unique\":true}],\"cacheEnabled\":false,\"deleted\":false,\"statusString\":\"loaded\",\"isSmart\":false,\"numberOfShards\":1,\"minReplicationFactor\":1,\"id\":\"10001\",\"name\":\"_users\",\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10007\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10018\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_analyzers\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10007\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10003\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10014\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_graphs\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10003\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10006\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10017\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_statisticsRaw\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10006\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"time\"],\"id\":\"10028\",\"inBackground\":false,\"name\":\"idx_1718347303770652672\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10012\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10023\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_appbundles\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10012\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10010\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10021\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_jobs\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10010\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"queue\",\"status\",\"delayUntil\"],\"id\":\"10030\",\"inBackground\":false,\"name\":\"idx_1718347303839858688\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"status\",\"queue\",\"delayUntil\"],\"id\":\"10031\",\"inBackground\":false,\"name\":\"idx_1718347303866073088\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10004\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10015\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_statistics\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10004\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"time\"],\"id\":\"10026\",\"inBackground\":false,\"name\":\"idx_1718347303708786688\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10005\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10016\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_statistics15\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10005\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false},{\"deduplicate\":true,\"estimates\":false,\"fields\":[\"time\"],\"id\":\"10027\",\"inBackground\":false,\"name\":\"idx_1718347303741292544\",\"sparse\":false,\"type\":\"skiplist\",\"unique\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10009\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10020\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_queues\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10009\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}},\"10013\":{\"usesRevisionsAsDocumentIds\":true,\"syncByRevision\":true,\"isSmartChild\":false,\"distributeShardsLike\":\"10001\",\"shardingStrategy\":\"hash\",\"shards\":{\"s10024\":[\"PRMR-igofehwp\",\"PRMR-lamgjtvh\"]},\"type\":2,\"status\":3,\"replicationFactor\":2,\"writeConcern\":1,\"name\":\"_frontend\",\"statusString\":\"loaded\",\"isSmart\":false,\"schema\":null,\"cacheEnabled\":false,\"numberOfShards\":1,\"id\":\"10013\",\"minReplicationFactor\":1,\"deleted\":false,\"shardKeys\":[\"_key\"],\"indexes\":[{\"id\":\"0\",\"type\":\"primary\",\"name\":\"primary\",\"fields\":[\"_key\"],\"unique\":true,\"sparse\":false}],\"isDisjoint\":false,\"waitForSync\":false,\"isSystem\":true,\"keyOptions\":{\"allowUserKeys\":true,\"type\":\"traditional\"}}}}}}}]"
	var s StateRoots

	require.NoError(t, json.Unmarshal([]byte(data), &s))

	t.Logf("%+v", s)
}

func Test_IsDBServerInSync(t *testing.T) {
	var state = GenerateState(t, NewDatabaseRandomGenerator().RandomCollection().
		WithShard().WithPlan("A", "B").WithCurrent("B", "C").Add().
		WithShard().WithPlan("A", "B", "C").WithCurrent("A", "B", "C").Add().
		WithWriteConcern(1).Add().Add(),
	)
	var expected = map[string]bool{
		"A": false, // in plan, not synced
		"B": true,  // in plan, synced
		"C": true,  // not in plan, synced
		"D": true,  // not in plan, not synced
	}
	for serverID, inSync := range expected {
		require.Equalf(t, inSync, state.IsDBServerInSync(serverID), "server %s", serverID)
	}
}

func Test_IsDBServerReadyToRestart(t *testing.T) {
	type testCase struct {
		generator StateGenerator
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
				require.Truef(t, s.IsDBServerReadyToRestart(server), "server %s should be able to restart", server)
			}
			for _, server := range tc.notReady {
				require.Falsef(t, s.IsDBServerReadyToRestart(server), "server %s should not be able to restart", server)
			}
		})
	}
}
