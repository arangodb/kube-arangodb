//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package tests

import (
	"context"
	"testing"

	"github.com/dchest/uniuri"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// TestScaleCluster tests scaling up/down the number of DBServers & coordinators
// of a cluster.
func TestScaleClusterNonTLS(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-scale-non-tls" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.TLS = api.TLSSpec{CASecretName: util.NewString("None")}
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Add 2 DBServers, 1 coordinator
	updated, err := updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = util.NewInt(5)
		spec.Coordinators.Count = util.NewInt(4)
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-up, in expected health in time: %v", err)
	}

	// Remove 3 DBServers, 2 coordinator
	updated, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = util.NewInt(3)
		spec.Coordinators.Count = util.NewInt(2)
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-down, in expected health in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}

func TestScaleCluster(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-scale" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.TLS = api.TLSSpec{}         // should auto-generate cert
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Add 2 DBServers, 1 coordinator
	updated, err := updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = util.NewInt(5)
		spec.Coordinators.Count = util.NewInt(4)
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-up, in expected health in time: %v", err)
	}

	// Remove 3 DBServers, 2 coordinator
	updated, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = util.NewInt(3)
		spec.Coordinators.Count = util.NewInt(2)
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-down, in expected health in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}

// TestScaleClusterWithSync tests scaling a cluster deployment with sync enabled.
func TestScaleClusterWithSync(t *testing.T) {
	longOrSkip(t)
	img := getEnterpriseImageOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-scale-sync" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Image = util.NewString(img)
	depl.Spec.Sync.Enabled = util.NewBool(true)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Create a syncmaster client
	syncClient := mustNewArangoSyncClient(ctx, kubecli, apiObject, t)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Wait for syncmasters to be available
	if err := waitUntilSyncVersionUp(syncClient, nil); err != nil {
		t.Fatalf("SyncMasters not running returning version in time: %v", err)
	}

	// Add 1 DBServer, 2 SyncMasters, 1 syncworker
	updated, err := updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = util.NewInt(4)
		spec.SyncMasters.Count = util.NewInt(5)
		spec.SyncWorkers.Count = util.NewInt(4)
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-up, in expected health in time: %v", err)
	}
	// Check number of syncmasters
	if err := waitUntilSyncMasterCountReached(syncClient, updated.Spec.SyncMasters.GetCount()); err != nil {
		t.Fatalf("Unexpected #syncmasters, after scale-up: %v", err)
	}
	// Check number of syncworkers
	if err := waitUntilSyncWorkerCountReached(syncClient, updated.Spec.SyncWorkers.GetCount()); err != nil {
		t.Fatalf("Unexpected #syncworkers, after scale-up: %v", err)
	}

	// Remove 1 DBServer, 2 SyncMasters & 1 SyncWorker
	updated, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = util.NewInt(3)
		spec.SyncMasters.Count = util.NewInt(3)
		spec.SyncWorkers.Count = util.NewInt(3)
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-down, in expected health in time: %v", err)
	}
	// Check number of syncmasters
	if err := waitUntilSyncMasterCountReached(syncClient, updated.Spec.SyncMasters.GetCount()); err != nil {
		t.Fatalf("Unexpected #syncmasters, after scale-up: %v", err)
	}
	// Check number of syncworkers
	if err := waitUntilSyncWorkerCountReached(syncClient, updated.Spec.SyncWorkers.GetCount()); err != nil {
		t.Fatalf("Unexpected #syncworkers, after scale-up: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
