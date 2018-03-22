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

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
)

// environements: provided from outside

func TestDeploymentSingleMMFiles(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeSingle, api.StorageEngineMMFiles)
}

func TestDeploymentSingleRocksDB(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeSingle, api.StorageEngineRocksDB)
}

func TestDeploymentClusterMMFiles(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeCluster, api.StorageEngineMMFiles)
}

func TestDeploymentClusterRocksDB(t *testing.T) {
	deploymentSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB)
}

func deploymentSubTest(t *testing.T, mode api.DeploymentMode, engine api.StorageEngine) error {
	// check environment
	longOrSkip(t)

	k8sNameSpace := getNamespace(t)
	k8sClient := mustNewKubeClient(t)
	deploymentClient := kubeArangoClient.MustNewInCluster() //

	// Prepare deployment config
	deploymentTemplate := newDeployment("test-1-deployment-" + string(mode) + "-" + string(engine) + "-" + uniuri.NewLen(4))
	deploymentTemplate.Spec.Mode = mode
	deploymentTemplate.Spec.StorageEngine = engine
	deploymentTemplate.Spec.TLS = api.TLSSpec{}                       // should auto-generate cert
	deploymentTemplate.Spec.SetDefaults(deploymentTemplate.GetName()) // this must be last

	// Create deployment
	deployment, err := deploymentClient.DatabaseV1alpha().ArangoDeployments(k8sNameSpace).Create(deploymentTemplate)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace, deploymentHasState(api.DeploymentStateRunning)); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	DBClient := mustNewArangodDatabaseClient(ctx, k8sClient, deployment, t)

	// deployment checks
	if deployment.Spec.Mode == api.DeploymentModeCluster {
		// Wait for cluster to be completely ready
		if err := waitUntilClusterHealth(DBClient, func(h driver.ClusterHealth) error {
			return clusterHealthEqualsSpec(h, deployment.Spec)
		}); err != nil {
			t.Fatalf("Cluster not running in expected health in time: %v", err)
		}
	} else if deployment.Spec.Mode == api.DeploymentModeSingle {
		if err := waitUntilVersionUp(DBClient); err != nil {
			t.Fatalf("Single Server not running in time: %v", err)
		}
	} else {
		t.Fatalf("DeploymentMode %v is not supported!", deployment.Spec.Mode)
	}

	// Cleanup
	removeDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace)

	return nil
}
