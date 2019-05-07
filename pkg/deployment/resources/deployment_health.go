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

package resources

import (
	"context"
	"fmt"
	"time"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
)

var (
	deploymentHealthFetchesCounters = metrics.MustRegisterCounterVec(metricsComponent, "deployment_health_fetches", "Number of times the health of the deployment was fetched", metrics.DeploymentName, metrics.Result)
	deploymentSyncFetchesCounters   = metrics.MustRegisterCounterVec(metricsComponent, "deployment_sync_fetches", "Number of times the sync status of shards of the deplyoment was fetched", metrics.DeploymentName, metrics.Result)
)

// RunDeploymentHealthLoop creates a loop to fetch the health of the deployment.
// The loop ends when the given channel is closed.
func (r *Resources) RunDeploymentHealthLoop(stopCh <-chan struct{}) {
	log := r.log
	deploymentName := r.context.GetAPIObject().GetName()

	if r.context.GetSpec().GetMode() != api.DeploymentModeCluster {
		// Deployment health is currently only applicable for clusters
		return
	}

	for {
		if err := r.fetchDeploymentHealth(); err != nil {
			log.Debug().Err(err).Msg("Failed to fetch deployment health")
			deploymentHealthFetchesCounters.WithLabelValues(deploymentName, metrics.Failed).Inc()
		} else {
			deploymentHealthFetchesCounters.WithLabelValues(deploymentName, metrics.Success).Inc()
		}
		select {
		case <-r.shardSync.triggerSyncInspection.Done():
		case <-time.After(time.Second * 5):
			// Continue
		case <-stopCh:
			// We're done
			return
		}
	}
}

// fetchDeploymentHealth performs a single fetch of cluster-health
// and stores it in-memory.
func (r *Resources) fetchDeploymentHealth() error {
	// Ask cluster for its health
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	client, err := r.context.GetDatabaseClient(ctx)
	if err != nil {
		return maskAny(err)
	}
	c, err := client.Cluster(ctx)
	if err != nil {
		return maskAny(err)
	}
	h, err := c.Health(ctx)
	if err != nil {
		return maskAny(err)
	}

	// Save cluster health
	r.health.mutex.Lock()
	defer r.health.mutex.Unlock()
	r.health.clusterHealth = h
	r.health.timestamp = time.Now()
	return nil
}

// GetDeploymentHealth returns a copy of the latest known state of cluster health
func (r *Resources) GetDeploymentHealth() (driver.ClusterHealth, error) {

	r.health.mutex.Lock()
	defer r.health.mutex.Unlock()
	if r.health.timestamp.IsZero() {
		return driver.ClusterHealth{}, fmt.Errorf("No cluster health available")
	}

	newhealth := r.health.clusterHealth
	newhealth.Health = make(map[driver.ServerID]driver.ServerHealth)

	for k, v := range r.health.clusterHealth.Health {
		newhealth.Health[k] = v
	}
	return newhealth, nil
}

// RunDeploymentShardSyncLoop creates a loop to fetch the sync status of shards of the deployment.
// The loop ends when the given channel is closed.
func (r *Resources) RunDeploymentShardSyncLoop(stopCh <-chan struct{}) {
	log := r.log
	deploymentName := r.context.GetAPIObject().GetName()

	if r.context.GetSpec().GetMode() != api.DeploymentModeCluster {
		// Deployment health is currently only applicable for clusters
		return
	}

	for {
		if err := r.fetchClusterShardSyncState(); err != nil {
			log.Debug().Err(err).Msg("Failed to fetch deployment shard sync state")
			deploymentSyncFetchesCounters.WithLabelValues(deploymentName, metrics.Failed).Inc()
		} else {
			deploymentSyncFetchesCounters.WithLabelValues(deploymentName, metrics.Success).Inc()
		}
		select {
		case <-time.After(time.Second * 30):
			// Continue
		case <-stopCh:
			// We're done
			return
		}
	}
}

// InvalidateSyncStatus resets the sync state to false and triggers an inspection
func (r *Resources) InvalidateSyncStatus() {
	r.log.Debug().Msg("Invalidating sync status due to previous events")
	r.shardSync.mutex.Lock()
	defer r.shardSync.mutex.Unlock()
	r.shardSync.allInSync = false
	r.shardSync.triggerSyncInspection.Trigger()
}

// fetchClusterShardSyncState performs a single fetch of the cluster inventory and
// checks if all shards are in sync
func (r *Resources) fetchClusterShardSyncState() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	c, err := r.context.GetDatabaseClient(ctx)
	if err != nil {
		return err
	}
	cluster, err := c.Cluster(ctx)
	if err != nil {
		return err
	}
	dbs, err := c.Databases(ctx)
	if err != nil {
		return err
	}

	allInSync := true
dbloop:
	for _, db := range dbs {
		inv, err := cluster.DatabaseInventory(ctx, db)
		if err != nil {
			return err
		}

		for _, col := range inv.Collections {
			if !col.AllInSync {
				r.log.Debug().Str("db", db.Name()).Str("col", col.Parameters.Name).Msg("Collection not in sync")
				allInSync = false
				break dbloop
			}
		}
	}

	r.shardSync.mutex.Lock()
	oldSyncState := r.shardSync.allInSync
	r.shardSync.allInSync = allInSync
	r.shardSync.timestamp = time.Now()
	r.shardSync.mutex.Unlock()

	if !oldSyncState && allInSync {
		r.log.Debug().Msg("Everything is in sync by now")
	}

	return nil
}

// GetShardSyncStatus returns true if all shards are in sync
func (r *Resources) GetShardSyncStatus() bool {
	if r.context.GetSpec().GetMode() != api.DeploymentModeCluster {
		// Shard sync status is only applicable for clusters
		return true
	}
	r.shardSync.mutex.Lock()
	defer r.shardSync.mutex.Unlock()
	return r.shardSync.allInSync
}
