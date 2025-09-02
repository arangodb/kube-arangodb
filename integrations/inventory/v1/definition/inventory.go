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
	"strconv"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

const (
	forceOneShardFlag = "--cluster.force-one-shard"
)

func (s *Inventory) JSON() ([]byte, error) {
	if s == nil {
		return []byte("{}"), nil
	}

	return ugrpc.Marshal(s, ugrpc.WithUseProtoNames(true))
}

func NewArangoDBConfiguration(spec api.DeploymentSpec, status api.DeploymentStatus) *ArangoDBConfiguration {
	var cfg ArangoDBConfiguration

	switch spec.Mode.Get() {
	case api.DeploymentModeSingle:
		cfg.Mode = ArangoDBMode_Single
	case api.DeploymentModeActiveFailover:
		cfg.Mode = ArangoDBMode_ActiveFailover
	case api.DeploymentModeCluster:
		cfg.Mode = ArangoDBMode_Cluster
	}

	cfg.Edition = ArangoDBEdition_Community

	if i := status.CurrentImage; i != nil {
		if i.Enterprise {
			cfg.Edition = ArangoDBEdition_Enterprise
		}

		cfg.Version = string(i.ArangoDBVersion)
	}

	if spec.GetMode() != api.DeploymentModeCluster {
		cfg.Sharding = ArangoDBSharding_Sharded
	} else {
		cfg.Sharding = getShardingFromArgs(spec.Coordinators.Args...)
	}

	return &cfg
}

func getShardingFromArgs(args ...string) ArangoDBSharding {
	for _, arg := range args {
		if arg == forceOneShardFlag {
			return ArangoDBSharding_OneShardEnforced
		}

		if v := strings.SplitN(arg, "=", 2); v[0] == forceOneShardFlag && len(v) == 2 {
			if q, err := strconv.ParseBool(v[1]); err != nil {
				continue
			} else if q {
				return ArangoDBSharding_OneShardEnforced
			}
		}
	}

	return ArangoDBSharding_Sharded
}
