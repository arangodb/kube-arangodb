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

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/dchest/uniuri"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// TestSyncToggleEnabled tests a normal cluster and enables sync later.
// Once sync is active, it is disabled again.
func TestSyncToggleEnabled(t *testing.T) {
	longOrSkip(t)
	img := getEnterpriseImageOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-sync-toggle-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Image = util.NewString(img)

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
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

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Enable sync
	updated, err := updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.Sync.Enabled = util.NewBool(true)
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait until sync jwt secret has been created
	if _, err := waitUntilSecret(kubecli, updated.Spec.Sync.Authentication.GetJWTSecretName(), ns, nil, deploymentReadyTimeout); err != nil {
		t.Fatalf("Sync JWT secret not created in time: %v", err)
	}

	// Create a syncmaster client
	syncClient := mustNewArangoSyncClient(ctx, kubecli, apiObject, t)

	// Wait for syncmasters to be available
	if err := waitUntilSyncVersionUp(syncClient, nil); err != nil {
		t.Fatalf("SyncMasters not running returning version in time: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-up, in expected health in time: %v", err)
	}
	// Check number of syncmasters
	if err := waitUntilSyncMasterCountReached(syncClient, 3); err != nil {
		t.Fatalf("Unexpected #syncmasters, after enabling sync: %v", err)
	}
	// Check number of syncworkers
	if err := waitUntilSyncWorkerCountReached(syncClient, 3); err != nil {
		t.Fatalf("Unexpected #syncworkers, after enabling sync: %v", err)
	}

	// Disable sync
	updated, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.Sync.Enabled = util.NewBool(false)
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for deployment to have no more syncmasters & workers
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, func(apiObject *api.ArangoDeployment) error {
		if cnt := len(apiObject.Status.Members.SyncMasters); cnt > 0 {
			return maskAny(fmt.Errorf("Expected 0 syncmasters, got %d", cnt))
		}
		if cnt := len(apiObject.Status.Members.SyncWorkers); cnt > 0 {
			return maskAny(fmt.Errorf("Expected 0 syncworkers, got %d", cnt))
		}
		return nil
	}); err != nil {
		t.Fatalf("Failed to reach deployment state without syncmasters & syncworkers: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
