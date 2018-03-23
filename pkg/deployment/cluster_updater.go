//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package deployment

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

// updateClusterServerCount updates the intended number of servers of the cluster.
func (d *Deployment) updateClusterServerCount(ctx context.Context) error {
	log := d.deps.Log
	c, err := d.clientCache.GetDatabase(ctx)
	if err != nil {
		return maskAny(err)
	}
	spec := d.apiObject.Spec
	coordinatorCount := spec.Coordinators.Count
	dbserverCount := spec.DBServers.Count
	if err := arangod.SetNumberOfServers(ctx, c.Connection(), coordinatorCount, dbserverCount); err != nil {
		log.Debug().Err(err).Msg("Failed to set number of servers")
		return maskAny(err)
	}
	d.lastNumberOfServers.mutex.Lock()
	defer d.lastNumberOfServers.mutex.Unlock()
	d.lastNumberOfServers.Coordinators = &coordinatorCount
	d.lastNumberOfServers.DBServers = &dbserverCount
	return nil
}
