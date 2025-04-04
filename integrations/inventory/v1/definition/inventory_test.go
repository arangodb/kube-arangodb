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

package definition

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

func Test_State_Marshal(t *testing.T) {
	s := Inventory{
		Configuration: &InventoryConfiguration{
			Hash: "xyz",
		},
		Arangodb: &ArangoDBConfiguration{
			Mode:     ArangoDBMode_Cluster,
			Edition:  ArangoDBEdition_Enterprise,
			Version:  "1.2.3",
			Sharding: ArangoDBSharding_Sharded,
		},
	}

	data, err := ugrpc.Marshal(&s, func(in *protojson.MarshalOptions) {
		in.EmitDefaultValues = true
	})
	require.NoError(t, err)

	t.Log(string(data))

	res, err := ugrpc.Unmarshal[*Inventory](data)
	require.NoError(t, err)

	require.NotNil(t, res)

	require.EqualValues(t, &s, res)
}

func Test_getShardingFromArgs(t *testing.T) {
	require.EqualValues(t, ArangoDBSharding_Sharded, getShardingFromArgs())
	require.EqualValues(t, ArangoDBSharding_OneShardEnforced, getShardingFromArgs(forceOneShardFlag))
	require.EqualValues(t, ArangoDBSharding_OneShardEnforced, getShardingFromArgs("--test=el", forceOneShardFlag))
	require.EqualValues(t, ArangoDBSharding_OneShardEnforced, getShardingFromArgs(fmt.Sprintf("%s=true", forceOneShardFlag)))
	require.EqualValues(t, ArangoDBSharding_Sharded, getShardingFromArgs(fmt.Sprintf("%s=false", forceOneShardFlag)))
	require.EqualValues(t, ArangoDBSharding_OneShardEnforced, getShardingFromArgs(fmt.Sprintf("%s=True", forceOneShardFlag)))
	require.EqualValues(t, ArangoDBSharding_Sharded, getShardingFromArgs(fmt.Sprintf("%s=Unknown", forceOneShardFlag)))
}
