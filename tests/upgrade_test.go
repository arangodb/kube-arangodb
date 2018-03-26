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
	arangod "github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

// TODO - environements (provided from outside)

// test upgrade single server mmfiles 3.2 -> 3.3
func TestUpgradeSingleMMFiles32to33(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeSingle, api.StorageEngineMMFiles, "3.2.0", "3.3.3")
}

// test upgrade single server rocksdb 3.3 -> 3.4
func TestUpgradeSingleRocksDB33to34(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeSingle, api.StorageEngineRocksDB, "3.3.0", "3.4")
}

// test upgrade resilient single server rocksdb 3.2 -> 3.3
func TestUpgradeResilientSingleRocksDB32to33(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeResilientSingle, api.StorageEngineRocksDB, "3.2.0", "3.3.3")
}

// test upgrade resilient single server mmfiles 3.3 -> 3.4
func TestUpgradeResilientSingleMMFiles33to34(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeResilientSingle, api.StorageEngineMMFiles, "3.3.0", "3.4")
}

// test upgrade cluster rocksdb 3.2 -> 3.3
func TestUpgradeClusterRocksDB32to33(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB, "3.2.0", "3.3.3")
}

// test upgrade cluster mmfiles 3.3 -> 3.4
func TestUpgradeClusterMMFiles33to34(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB, "3.3.0", "3.4")
}

func upgradeSubTest(t *testing.T, mode api.DeploymentMode, engine api.StorageEngine, fromVersion, toVersion string) error {
	// check environment
	longOrSkip(t)

	k8sNameSpace := getNamespace(t)
	k8sClient := mustNewKubeClient(t)
	deploymentClient := kubeArangoClient.MustNewInCluster()

	// Prepare deployment config
	deploymentTemplate := newDeployment("test-1-deployment-" + string(mode) + "-" + string(engine) + "-" + uniuri.NewLen(4))
	deploymentTemplate.Spec.Mode = api.NewMode(mode)
	deploymentTemplate.Spec.StorageEngine = api.NewStorageEngine(engine)
	deploymentTemplate.Spec.TLS = api.TLSSpec{} // should auto-generate cert
	deploymentTemplate.Spec.Image = "arangodb/arangodb:" + fromVersion
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

	// Try to change image version
	updated, err = updateDeployment(c, depl.GetName(), ns,
		func(spec *api.DeploymentSpec) {
			spec.Image = "arangodb/arangodb:" + toVersion
		})
	if err != nil {
		t.Fatalf("Failed to upgrade the Image from version : " + fromVersion + " to version: " + toVersion)
	}

	// Create a database client
	ctx := context.Background()
	DBClient := mustNewArangodDatabaseClient(ctx, k8sClient, deployment, t)

	// deployment checks
	switch mode := deployment.Spec.GetMode(); mode {
	case api.DeploymentModeCluster:
		// Wait for cluster to be completely ready
		if err := waitUntilClusterHealth(DBClient, func(h driver.ClusterHealth) error {
			return clusterHealthEqualsSpec(h, deployment.Spec)
		}); err != nil {
			t.Fatalf("Cluster not running in expected health in time: %v", err)
		}
	case api.DeploymentModeSingle:
		if err := waitUntilVersionUp(DBClient); err != nil {
			t.Fatalf("Single Server not running in time: %v", err)
		}
	case api.DeploymentModeResilientSingle:
		if err := waitUntilVersionUp(DBClient); err != nil {
			t.Fatalf("Single Server not running in time: %v", err)
		}

		members := deployment.Status.Members
		singles := members.Single
		agents := members.Agents

		if len(singles) != 2 || len(agents) != 3 {
			t.Fatal("Wrong number of servers: single %v - agents %v", len(singles), len(agents))
		}

		for _, agent := range agents {
			dbclient, err := arangod.CreateArangodClient(ctx, k8sClient.CoreV1(), deployment, api.ServerGroupAgents, agent.ID)
			if err != nil {
				t.Fatal("Unable to create connection to: %v", agent.ID)
			}
			if waitUntilVersionUp(dbclient) != nil {
				t.Fatal("Version check failed for: %v", single.ID)
			}
		}
		for _, single := range singles {
			dbclient, err := arangod.CreateArangodClient(ctx, k8sClient.CoreV1(), deployment, api.ServerGroupAgents, single.ID)
			if err != nil {
				t.Fatal("Unable to create connection to: %v", single.ID)
			}
			if waitUntilVersionUp(dbclient) != nil {
				t.Fatal("Version check failed for: %v", single.ID)
			}
		}
	default:
		t.Fatalf("DeploymentMode %v is not supported!", mode)
	}

	// Cleanup
	removeDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace)

	return nil
}
