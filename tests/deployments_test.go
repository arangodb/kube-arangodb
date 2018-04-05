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

// test deployment resilient single server mmfiles
func TestDeploymentResilientSingleMMFiles(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeResilientSingle, api.StorageEngineMMFiles)
}

// test deployment resilient single server rocksdb
func TestDeploymentResilientSingleRocksDB(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeResilientSingle, api.StorageEngineRocksDB)
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

	k8sNameSpace := getNamespace(t)
	k8sClient := mustNewKubeClient(t)
	deploymentClient := kubeArangoClient.MustNewInCluster()

	// Prepare deployment config
	deploymentTemplate := newDeployment("test-1-deployment-" + string(mode) + "-" + string(engine) + "-" + uniuri.NewLen(4))
	deploymentTemplate.Spec.Mode = api.NewMode(mode)
	deploymentTemplate.Spec.StorageEngine = api.NewStorageEngine(engine)
	deploymentTemplate.Spec.TLS = api.TLSSpec{}                       // should auto-generate cert
	deploymentTemplate.Spec.SetDefaults(deploymentTemplate.GetName()) // this must be last

	// Create deployment
	_, err := deploymentClient.DatabaseV1alpha().ArangoDeployments(k8sNameSpace).Create(deploymentTemplate)
	require.NoError(t, err, fmt.Sprintf("Create deployment failed: %v", err))

	// Wait for deployment to be ready
	deployment, err := waitUntilDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment not running in time: %v", err))

	// Create a database client
	ctx := context.Background()
	DBClient := mustNewArangodDatabaseClient(ctx, k8sClient, deployment, t)
	require.NoError(t, waitUntilArangoDeploymentHealthy(deployment, DBClient, k8sClient, ""), fmt.Sprintf("Deployment not healthy in time: %v", err))

	// Cleanup
	removeDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace)

	return nil
}

// test a setup containing multiple deployments
func TestMultiDeployment(t *testing.T) {
	longOrSkip(t)

	k8sNameSpace := getNamespace(t)
	k8sClient := mustNewKubeClient(t)
	deploymentClient := kubeArangoClient.MustNewInCluster()

	// Prepare deployment configurations
	deploymentTemplate1 := newDeployment("test-multidep1-1-" + uniuri.NewLen(4))
	deploymentTemplate1.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	deploymentTemplate1.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineRocksDB)
	deploymentTemplate1.Spec.TLS = api.TLSSpec{}                        // should auto-generate cert
	deploymentTemplate1.Spec.SetDefaults(deploymentTemplate1.GetName()) // this must be last

	deploymentTemplate2 := newDeployment("test-multidep1-2-" + uniuri.NewLen(4))
	deploymentTemplate2.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	deploymentTemplate2.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineMMFiles)
	deploymentTemplate2.Spec.TLS = api.TLSSpec{}                        // should auto-generate cert
	deploymentTemplate2.Spec.SetDefaults(deploymentTemplate2.GetName()) // this must be last

	// Create deployments
	_, err := deploymentClient.DatabaseV1alpha().ArangoDeployments(k8sNameSpace).Create(deploymentTemplate1)
	require.NoError(t, err, fmt.Sprintf("Deployment creation failed: %v", err))

	_, err = deploymentClient.DatabaseV1alpha().ArangoDeployments(k8sNameSpace).Create(deploymentTemplate2)
	require.NoError(t, err, fmt.Sprintf("Deployment creation failed: %v", err))

	// Wait for deployments to be ready
	deployment1, err := waitUntilDeployment(deploymentClient, deploymentTemplate1.GetName(), k8sNameSpace, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment not running in time: %v", err))

	deployment2, err := waitUntilDeployment(deploymentClient, deploymentTemplate2.GetName(), k8sNameSpace, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment not running in time: %v", err))

	require.True(t, deployment1 != nil && deployment2 != nil, "deployment is nil")

	// Create a database clients
	ctx := context.Background()
	DBClient1 := mustNewArangodDatabaseClient(ctx, k8sClient, deployment1, t)
	require.NoError(t, waitUntilArangoDeploymentHealthy(deployment1, DBClient1, k8sClient, ""), fmt.Sprintf("Deployment not healthy in time: %v", err))
	DBClient2 := mustNewArangodDatabaseClient(ctx, k8sClient, deployment2, t)
	require.NoError(t, waitUntilArangoDeploymentHealthy(deployment1, DBClient1, k8sClient, ""), fmt.Sprintf("Deployment not healthy in time: %v", err))

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
	removeDeployment(deploymentClient, deploymentTemplate1.GetName(), k8sNameSpace)
	removeDeployment(deploymentClient, deploymentTemplate2.GetName(), k8sNameSpace)

}

func containsCollection(colls []driver.Collection, name string) bool {
	for _, col := range colls {
		if name == col.Name() {
			return true
		}
	}
	return false
}
