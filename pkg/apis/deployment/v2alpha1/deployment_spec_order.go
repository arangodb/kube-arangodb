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

package v2alpha1

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

type DeploymentSpecOrder string

func (d DeploymentSpecOrder) String() string {
	return string(d)
}

func (d *DeploymentSpecOrder) Equal(b *DeploymentSpecOrder) bool {
	if d == nil && b == nil {
		return true
	}

	if d == nil || b == nil {
		return false
	}

	return *d == *b
}

func (d *DeploymentSpecOrder) Validate() error {
	if d == nil {
		return nil
	}

	switch v := *d; v {
	case DeploymentSpecOrderStandard, DeploymentSpecOrderCoordinatorFirst:
		return nil
	default:
		return errors.Errorf("Invalid Order `%s`", v)
	}
}

func (d DeploymentSpecOrder) Groups() ServerGroups {
	switch d {
	case DeploymentSpecOrderCoordinatorFirst:
		return []ServerGroup{
			ServerGroupAgents,
			ServerGroupSingle,
			ServerGroupCoordinators,
			ServerGroupDBServers,
			ServerGroupSyncMasters,
			ServerGroupSyncWorkers,
			ServerGroupGateways,
		}
	default:
		return AllServerGroups
	}
}

const (
	DeploymentSpecOrderStandard         DeploymentSpecOrder = "standard"
	DeploymentSpecOrderCoordinatorFirst DeploymentSpecOrder = "coordinatorFirst"
)
