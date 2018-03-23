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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

// listenForClusterEvents keep listening for changes entered in the UI of the cluster.
func (d *Deployment) listenForClusterEvents(stopCh <-chan struct{}) {
	for {
		delay := time.Second * 2

		// Inspect once
		ctx := context.Background()
		if err := d.inspectCluster(ctx); err != nil {
			d.deps.Log.Debug().Err(err).Msg("Cluster inspection failed")
		}

		select {
		case <-time.After(delay):
			// Continue
		case <-stopCh:
			// We're done
			return
		}
	}
}

// Perform a single inspection of the cluster
func (d *Deployment) inspectCluster(ctx context.Context) error {
	log := d.deps.Log
	c, err := d.clientCache.GetDatabase(ctx)
	if err != nil {
		return maskAny(err)
	}
	req, err := arangod.GetNumberOfServers(ctx, c.Connection())
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get number of servers")
		return maskAny(err)
	}
	if req.Coordinators == nil && req.DBServers == nil {
		// Nothing to check
		return nil
	}
	coordinatorsChanged := false
	dbserversChanged := false
	d.lastNumberOfServers.mutex.Lock()
	defer d.lastNumberOfServers.mutex.Unlock()
	desired := d.lastNumberOfServers.NumberOfServers
	if req.Coordinators != nil && desired.Coordinators != nil && req.GetCoordinators() != desired.GetCoordinators() {
		// #Coordinator has changed
		coordinatorsChanged = true
	}
	if req.DBServers != nil && desired.DBServers != nil && req.GetDBServers() != desired.GetDBServers() {
		// #DBServers has changed
		dbserversChanged = true
	}
	if !coordinatorsChanged && !dbserversChanged {
		// Nothing has changed
		return nil
	}
	// Let's update the spec
	current, err := d.deps.DatabaseCRCli.DatabaseV1alpha().ArangoDeployments(d.apiObject.Namespace).Get(d.apiObject.Name, metav1.GetOptions{})
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get current deployment")
		return maskAny(err)
	}
	if coordinatorsChanged {
		current.Spec.Coordinators.Count = util.NewInt(req.GetCoordinators())
	}
	if dbserversChanged {
		current.Spec.DBServers.Count = util.NewInt(req.GetDBServers())
	}
	if err := d.updateCRSpec(current.Spec); err != nil {
		log.Warn().Err(err).Msg("Failed to update current deployment")
		return maskAny(err)
	}
	return nil
}
