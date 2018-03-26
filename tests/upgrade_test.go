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
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// TODO - environments (provided from outside)

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
	deploymentTemplate := newDeployment("test-upgrade-" + string(mode) + "-" + string(engine) + "-" + fromVersion + "to" + toVersion + "-" + uniuri.NewLen(4))
	deploymentTemplate.Spec.Mode = api.NewMode(mode)
	deploymentTemplate.Spec.StorageEngine = api.NewStorageEngine(engine)
	deploymentTemplate.Spec.TLS = api.TLSSpec{} // should auto-generate cert
	deploymentTemplate.Spec.Image = util.NewString("arangodb/arangodb:" + fromVersion)
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

	arangoDeploymentHealthy(t, deployment, k8sClient)

	// Try to change image version
	deployment, err = updateDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace,
		func(spec *api.DeploymentSpec) {
			spec.Image = util.NewString("arangodb/arangodb:" + toVersion)
		})
	if err != nil {
		t.Fatalf("Failed to upgrade the Image from version : " + fromVersion + " to version: " + toVersion)
	}

	arangoDeploymentHealthy(t, deployment, k8sClient)

	ctx := context.Background()
	DBClient := mustNewArangodDatabaseClient(ctx, k8sClient, deployment, t)
	DBClient.Version(ctx)
	if vInfo, err := DBClient.Version(ctx); err != nil {
		t.Fatalf("Failed to receive version: %v", err)
	} else if vInfo.Version.CompareTo(driver.Version(toVersion)) != 0 {
		t.Fatalf("version %v returned by _api/version does not match specified version %v", vInfo.Version, toVersion)
	}

	// Cleanup
	removeDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace)

	return nil
}
