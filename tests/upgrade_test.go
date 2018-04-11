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
	"strings"
	"testing"

	"github.com/dchest/uniuri"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// test upgrade single server mmfiles 3.2 -> 3.3
func TestUpgradeSingleMMFiles32to33(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeSingle, api.StorageEngineMMFiles, "3.2.12", "3.3.4")
}

// // test upgrade single server rocksdb 3.3 -> 3.4
// func TestUpgradeSingleRocksDB33to34(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeSingle, api.StorageEngineRocksDB, "3.3.4", "3.4.0")
// }

/*// test upgrade active-failover server rocksdb 3.3 -> 3.4
func TestUpgradeActiveFailoverRocksDB33to34(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeActiveFailover, api.StorageEngineRocksDB, "3.3.5", "3.4.0")
}*/

// // test upgrade active-failover server mmfiles 3.3 -> 3.4
// func TestUpgradeActiveFailoverMMFiles33to34(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeActiveFailover, api.StorageEngineMMFiles, "3.3.0", "3.4.0")
// }

// test upgrade cluster rocksdb 3.2 -> 3.3
func TestUpgradeClusterRocksDB32to33(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB, "3.2.12", "3.3.4")
}

// // test upgrade cluster mmfiles 3.3 -> 3.4
// func TestUpgradeClusterMMFiles33to34(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB, "3.3.4", "3.4.0")
// }

// test downgrade single server mmfiles 3.3.3 -> 3.3.2
func TestDowngradeSingleMMFiles333to332(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeSingle, api.StorageEngineMMFiles, "3.3.3", "3.3.2")
}

// test downgrade ActiveFailover server rocksdb 3.3.3 -> 3.3.2
func TestDowngradeActiveFailoverRocksDB333to332(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeActiveFailover, api.StorageEngineRocksDB, "3.3.3", "3.3.2")
}

// test downgrade cluster rocksdb 3.3.3 -> 3.3.2
func TestDowngradeClusterRocksDB332to332(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB, "3.3.3", "3.3.2")
}

func upgradeSubTest(t *testing.T, mode api.DeploymentMode, engine api.StorageEngine, fromVersion, toVersion string) error {
	// check environment
	longOrSkip(t)

	ns := getNamespace(t)
	kubecli := mustNewKubeClient(t)
	c := kubeArangoClient.MustNewInCluster()

	depl := newDeployment(strings.Replace(fmt.Sprintf("tu-%s-%s-%st%s-%s", mode[:2], engine[:2], fromVersion, toVersion, uniuri.NewLen(4)), ".", "", -1))
	depl.Spec.Mode = api.NewMode(mode)
	depl.Spec.StorageEngine = api.NewStorageEngine(engine)
	depl.Spec.TLS = api.TLSSpec{} // should auto-generate cert
	depl.Spec.Image = util.NewString("arangodb/arangodb:" + fromVersion)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	deployment, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	deployment, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	DBClient := mustNewArangodDatabaseClient(ctx, kubecli, deployment, t)

	if err := waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, ""); err != nil {
		t.Fatalf("Deployment not healthy in time: %v", err)
	}

	// Try to change image version
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(spec *api.DeploymentSpec) {
			spec.Image = util.NewString("arangodb/arangodb:" + toVersion)
		})
	if err != nil {
		t.Fatalf("Failed to upgrade the Image from version : " + fromVersion + " to version: " + toVersion)
	}

	deployment, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	if err := waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, toVersion); err != nil {
		t.Fatalf("Deployment not healthy in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)

	return nil
}
