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
	"testing"

	"github.com/dchest/uniuri"

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
	deployment, err := deploymentClient.DatabaseV1alpha().ArangoDeployments(k8sNameSpace).Create(deploymentTemplate)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	deployment, err = waitUntilDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace, deploymentHasState(api.DeploymentStateRunning))
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	DBClient := mustNewArangodDatabaseClient(ctx, k8sClient, deployment, t)

	if err := waitUntilArangoDeploymentHealthy(deployment, DBClient, k8sClient, ""); err != nil {
		t.Fatalf("Deployment not healthy in time: %v", err)
	}

	// Cleanup
	removeDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace)

	return nil
}
