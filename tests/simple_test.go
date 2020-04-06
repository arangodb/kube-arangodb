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
	"github.com/stretchr/testify/assert"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// TestSimpleSingle tests the creating of a single server deployment
// with default settings.
func TestSimpleSingle(t *testing.T) {
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-sng-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Check server role
	assert.NoError(t, testServerRole(ctx, client, driver.ServerRoleSingle))
}

// TestSimpleActiveFailover tests the creating of a ActiveFailover server deployment
// with default settings.
func TestSimpleActiveFailover(t *testing.T) {
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-rs-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeActiveFailover)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("ActiveFailover servers not running returning version in time: %v", err)
	}

	// Check server role
	assert.NoError(t, testServerRole(ctx, client, driver.ServerRoleSingleActive))
}

// TestSimpleCluster tests the creating of a cluster deployment
// with default settings.
func TestSimpleCluster(t *testing.T) {
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-cls-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Cluster not running returning version in time: %v", err)
	}

	// Check server role
	assert.NoError(t, testServerRole(ctx, client, driver.ServerRoleCoordinator))
}

// TestSimpleClusterWithSync tests the creating of a cluster deployment
// with default settings and sync enabled.
func TestSimpleClusterWithSync(t *testing.T) {
	img := getEnterpriseImageOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-cls-sync-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Image = util.NewString(img)
	depl.Spec.Sync.Enabled = util.NewBool(true)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Cluster not running returning version in time: %v", err)
	}

	// Create a syncmaster client
	syncClient := mustNewArangoSyncClient(ctx, kubecli, apiObject, t)

	// Wait for syncmasters to be available
	if err := waitUntilSyncVersionUp(syncClient, nil); err != nil {
		t.Fatalf("SyncMasters not running returning version in time: %v", err)
	}

	// Check server role
	assert.NoError(t, testServerRole(ctx, client, driver.ServerRoleCoordinator))
}
