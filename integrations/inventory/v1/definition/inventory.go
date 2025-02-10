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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

func (s *Inventory) JSON() ([]byte, error) {
	if s == nil {
		return []byte("{}"), nil
	}

	return ugrpc.Marshal(s)
}

func NewArangoDBConfiguration(spec api.DeploymentSpec, status api.DeploymentStatus) *ArangoDBConfiguration {
	var cfg ArangoDBConfiguration

	switch spec.Mode.Get() {
	case api.DeploymentModeSingle:
		cfg.Mode = ArangoDBMode_ARANGO_DB_MODE_SINGLE
	case api.DeploymentModeActiveFailover:
		cfg.Mode = ArangoDBMode_ARANGO_DB_MODE_ACTIVE_FAILOVER
	case api.DeploymentModeCluster:
		cfg.Mode = ArangoDBMode_ARANGO_DB_MODE_CLUSTER
	}

	cfg.Edition = ArangoDBEdition_ARANGO_DB_EDITION_COMMUNITY

	if i := status.CurrentImage; i != nil {
		if i.Enterprise {
			cfg.Edition = ArangoDBEdition_ARANGO_DB_EDITION_ENTERPRISE
		}

		cfg.Version = string(i.ArangoDBVersion)
	}

	return &cfg
}
