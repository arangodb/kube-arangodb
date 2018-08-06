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
// Author Jan Christoph Uhde <jan@uhdejc.com>
//
package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	driver "github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
)

// test deployment single server mmfiles
func TestDeploymentSingleMMFiles(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeSingle, api.StorageEngineMMFiles)
}

// test deployment single server rocksdb
func TestDeploymentSingleRocksDB(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeSingle, api.StorageEngineRocksDB)
}

// test deployment active-failover server mmfiles
func TestDeploymentActiveFailoverMMFiles(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeActiveFailover, api.StorageEngineMMFiles)
}

// test deployment active-failover server rocksdb
func TestDeploymentActiveFailoverRocksDB(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeActiveFailover, api.StorageEngineRocksDB)
}

// test deployment cluster mmfiles
func TestDeploymentClusterMMFiles(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeCluster, api.StorageEngineMMFiles)
}

// test deployment cluster rocksdb
func TestDeploymentClusterRocksDB(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB)
}

func deploymentSubTest(t *testing.T, mode api.DeploymentMode, engine api.StorageEngine) error {
	// check environment
	longOrSkip(t)

	ns := getNamespace(t)
	kubecli := mustNewKubeClient(t)
	c := kubeArangoClient.MustNewInCluster()

	// Prepare deployment config
	depl := newDeployment("test-deployment-" + string(mode) + "-" + string(engine) + "-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(mode)
	depl.Spec.StorageEngine = api.NewStorageEngine(engine)
	depl.Spec.TLS = api.TLSSpec{}         // should auto-generate cert
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	require.NoError(t, err, fmt.Sprintf("Create deployment failed: %v", err))
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	deployment, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment not running in time: %v", err))

	// Create a database client
	ctx := context.Background()
	DBClient := mustNewArangodDatabaseClient(ctx, kubecli, deployment, t, nil)
	require.NoError(t, waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, ""), fmt.Sprintf("Deployment not healthy in time: %v", err))

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)

	return nil
}

// test a setup containing multiple deployments
func TestMultiDeployment(t *testing.T) {
	longOrSkip(t)

	ns := getNamespace(t)
	kubecli := mustNewKubeClient(t)
	c := kubeArangoClient.MustNewInCluster()

	// Prepare deployment configurations
	depl1 := newDeployment("test-multidep-1-" + uniuri.NewLen(4))
	depl1.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl1.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineRocksDB)
	depl1.Spec.TLS = api.TLSSpec{}          // should auto-generate cert
	depl1.Spec.SetDefaults(depl1.GetName()) // this must be last

	depl2 := newDeployment("test-multidep-2-" + uniuri.NewLen(4))
	depl2.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	depl2.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineMMFiles)
	depl2.Spec.TLS = api.TLSSpec{}          // should auto-generate cert
	depl2.Spec.SetDefaults(depl2.GetName()) // this must be last

	// Create deployments
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl1)
	require.NoError(t, err, fmt.Sprintf("Deployment creation failed: %v", err))
	defer deferedCleanupDeployment(c, depl1.GetName(), ns)

	_, err = c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl2)
	require.NoError(t, err, fmt.Sprintf("Deployment creation failed: %v", err))
	defer deferedCleanupDeployment(c, depl2.GetName(), ns)

	// Wait for deployments to be ready
	deployment1, err := waitUntilDeployment(c, depl1.GetName(), ns, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment not running in time: %v", err))

	deployment2, err := waitUntilDeployment(c, depl2.GetName(), ns, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment not running in time: %v", err))

	require.True(t, deployment1 != nil && deployment2 != nil, "deployment is nil")

	// Create a database clients
	ctx := context.Background()
	DBClient1 := mustNewArangodDatabaseClient(ctx, kubecli, deployment1, t, nil)
	require.NoError(t, waitUntilArangoDeploymentHealthy(deployment1, DBClient1, kubecli, ""), fmt.Sprintf("Deployment not healthy in time: %v", err))
	DBClient2 := mustNewArangodDatabaseClient(ctx, kubecli, deployment2, t, nil)
	require.NoError(t, waitUntilArangoDeploymentHealthy(deployment1, DBClient1, kubecli, ""), fmt.Sprintf("Deployment not healthy in time: %v", err))

	// Test if we are able to create a collections in both deployments.
	db1, err := DBClient1.Database(ctx, "_system")
	require.NoError(t, err, "failed to get database")
	_, err = db1.CreateCollection(ctx, "col1", nil)
	require.NoError(t, err, "failed to create collection")

	db2, err := DBClient2.Database(ctx, "_system")
	require.NoError(t, err, "failed to get database")
	_, err = db2.CreateCollection(ctx, "col2", nil)
	require.NoError(t, err, "failed to create collection")

	// The newly created collections must be (only) visible in the deployment
	// that it was created in. The following lines ensure this behavior.
	collections1, err := db1.Collections(ctx)
	require.NoError(t, err, "failed to get collections")
	collections2, err := db2.Collections(ctx)
	require.NoError(t, err, "failed to get collections")

	assert.True(t, containsCollection(collections1, "col1"), "collection missing")
	assert.True(t, containsCollection(collections2, "col2"), "collection missing")
	assert.False(t, containsCollection(collections1, "col2"), "collection must not be in this deployment")
	assert.False(t, containsCollection(collections2, "col1"), "collection must not be in this deployment")

	// Cleanup
	removeDeployment(c, depl1.GetName(), ns)
	removeDeployment(c, depl2.GetName(), ns)

}

func containsCollection(colls []driver.Collection, name string) bool {
	for _, col := range colls {
		if name == col.Name() {
			return true
		}
	}
	return false
}
