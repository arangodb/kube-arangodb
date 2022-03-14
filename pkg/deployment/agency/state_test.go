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
	const dbName, colID, sh1, sh2 = "db1", "1001", "shard1", "shard2"
	const sid1, sid2, sid3, sid4 = "PRMR-1", "PRMR-2", "PRMR-3", "PRMR-4"
	var s = State{
		Supervision: StateSupervision{},
		Plan: StatePlan{
			Collections: map[string]StatePlanDBCollections{
				dbName: map[string]StatePlanCollection{
					colID: {
						Shards: map[string]ShardServers{
							sh1: []string{sid1, sid2},
							sh2: []string{sid1, sid2, sid3},
						},
					},
				},
			},
		},
		Current: StateCurrent{
			Collections: map[string]StateCurrentDBCollections{
				dbName: map[string]StateCurrentDBCollection{
					colID: map[string]StateCurrentDBShard{
						sh1: {Servers: []string{sid2, sid3}},
						sh2: {Servers: []string{sid1, sid2, sid3}},
					},
				},
			},
		},
	}
	var expected = map[string]bool{
		sid1: false, // in plan, not synced
		sid2: true,  // in plan, synced
		sid3: true,  // not in plan, synced
		sid4: true,  // not in plan, not synced
	}

	for serverID, inSync := range expected {
		require.Equalf(t, inSync, s.IsDBServerInSync(serverID), "server %s", serverID)
	}
}

func Test_IsDBServerIsReadyToRestart(t *testing.T) {
	const dbName, colID, sh1, sh2, sh3 = "db1", "1001", "shard1", "shard2", "shard3"
	const sid1, sid2, sid3, sid4 = "PRMR-1", "PRMR-2", "PRMR-3", "PRMR-4"
	var wc1, wc2, wc3 = 1, 2, 3

	var testCases = map[string]struct {
		state    State
		expected map[string]bool
	}{
		"write-concern-1": {
			state: State{
				Supervision: StateSupervision{},
				Plan: StatePlan{
					Collections: map[string]StatePlanDBCollections{
						dbName: map[string]StatePlanCollection{
							colID: {
								WriteConcern: &wc1,
								Shards: map[string]ShardServers{
									sh1: []string{sid1, sid2},
									sh2: []string{sid1, sid2, sid3},
								},
							},
						},
					},
				},
				Current: StateCurrent{
					Collections: map[string]StateCurrentDBCollections{
						dbName: map[string]StateCurrentDBCollection{
							colID: map[string]StateCurrentDBShard{
								sh1: {Servers: []string{sid2, sid3}},
								sh2: {Servers: []string{sid1, sid2, sid3}},
							},
						},
					},
				},
			},
			expected: map[string]bool{
				sid1: true,
				sid2: true,
				sid3: true,
				sid4: true,
			},
		},
		"write-concern-2": {
			state: State{
				Supervision: StateSupervision{},
				Plan: StatePlan{
					Collections: map[string]StatePlanDBCollections{
						dbName: map[string]StatePlanCollection{
							colID: {
								WriteConcern: &wc2,
								Shards: map[string]ShardServers{
									sh1: []string{sid1, sid2},
									sh2: []string{sid1, sid2, sid3},
									sh3: []string{sid1, sid3},
								},
							},
						},
					},
				},
				Current: StateCurrent{
					Collections: map[string]StateCurrentDBCollections{
						dbName: map[string]StateCurrentDBCollection{
							colID: map[string]StateCurrentDBShard{
								sh1: {Servers: []string{sid3}},
								sh2: {Servers: []string{sid1, sid2, sid3}},
								sh3: {Servers: []string{}},
							},
						},
					},
				},
			},
			expected: map[string]bool{
				sid1: false,
				sid2: false,
				sid3: true,
				sid4: true,
			},
		},
		"write-concern-3": {
			state: State{
				Supervision: StateSupervision{},
				Plan: StatePlan{
					Collections: map[string]StatePlanDBCollections{
						dbName: map[string]StatePlanCollection{
							colID: {
								WriteConcern: &wc3,
								Shards: map[string]ShardServers{
									sh1: []string{sid1, sid2, sid3},
									sh2: []string{sid1, sid2, sid3},
								},
							},
						},
					},
				},
				Current: StateCurrent{
					Collections: map[string]StateCurrentDBCollections{
						dbName: map[string]StateCurrentDBCollection{
							colID: map[string]StateCurrentDBShard{
								sh1: {Servers: []string{sid3}},
								sh2: {Servers: []string{sid1, sid2, sid3}},
							},
						},
					},
				},
			},
			expected: map[string]bool{
				sid1: false,
				sid2: false,
				sid3: true,
				sid4: true,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			for server, readyToRestart := range testCase.expected {
				require.Equalf(t, readyToRestart, testCase.state.IsDBServerReadyToRestart(server), "server %s", server)
			}
		})
	}

}
