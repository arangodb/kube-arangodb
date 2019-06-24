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

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/dchest/uniuri"
)

// func TestUpgradeClusterRocksDB33pto34p(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB, "arangodb/arangodb-preview:3.3", "arangodb/arangodb-preview:3.4")
// }

// test upgrade single server mmfiles 3.2 -> 3.3
// func TestUpgradeSingleMMFiles32to33(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeSingle, api.StorageEngineMMFiles, "arangodb/arangodb:3.2.16", "arangodb/arangodb:3.3.13")
// }

// // test upgrade single server rocksdb 3.3 -> 3.4
// func TestUpgradeSingleRocksDB33to34(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeSingle, api.StorageEngineRocksDB, "3.3.13", "3.4.0")
// }

/*// test upgrade active-failover server rocksdb 3.3 -> 3.4
func TestUpgradeActiveFailoverRocksDB33to34(t *testing.T) {
	upgradeSubTest(t, api.DeploymentModeActiveFailover, api.StorageEngineRocksDB, "3.3.13", "3.4.0")
}*/

// // test upgrade active-failover server mmfiles 3.3 -> 3.4
// func TestUpgradeActiveFailoverMMFiles33to34(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeActiveFailover, api.StorageEngineMMFiles, "3.3.13", "3.4.0")
// }

// test upgrade cluster rocksdb 3.2 -> 3.3
// func TestUpgradeClusterRocksDB32to33(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB, "3.2.16", "3.3.13")
// }

// // test upgrade cluster mmfiles 3.3 -> 3.4
// func TestUpgradeClusterMMFiles33to34(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB, "3.3.13", "3.4.0")
// }

// // test downgrade single server mmfiles 3.3.17 -> 3.3.16
// func TestDowngradeSingleMMFiles3317to3316(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeSingle, api.StorageEngineMMFiles, "arangodb/arangodb:3.3.16", "arangodb/arangodb:3.3.17")
// }

// // test downgrade ActiveFailover server rocksdb 3.3.17 -> 3.3.16
// func TestDowngradeActiveFailoverRocksDB3317to3316(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeActiveFailover, api.StorageEngineRocksDB, "arangodb/arangodb:3.3.16", "arangodb/arangodb:3.3.17")
// }

// // test downgrade cluster rocksdb 3.3.17 -> 3.3.16
// func TestDowngradeClusterRocksDB3317to3316(t *testing.T) {
// 	upgradeSubTest(t, api.DeploymentModeCluster, api.StorageEngineRocksDB, "arangodb/arangodb:3.3.16", "arangodb/arangodb:3.3.17")
// }

func TestUpgradeClusterRocksDB3322Cto342C(t *testing.T) {
	runUpgradeTest(t, &upgradeTest{
		fromVersion: "3.3.22",
		toVersion:   "3.4.6.1",
		shortTest:   true,
	})
}

func TestUpgradeClusterRocksDB3316Cto3323C(t *testing.T) {
	runUpgradeTest(t, &upgradeTest{
		fromVersion: "3.3.16",
		toVersion:   "3.3.23",
		shortTest:   false,
	})
}

func TestUpgradeClusterRocksDB346Cto3461C(t *testing.T) {
	runUpgradeTest(t, &upgradeTest{
		fromVersion: "3.4.6",
		toVersion:   "3.4.6.1",
		shortTest:   true,
	})
}

type upgradeTest struct {
	fromVersion string
	toVersion   string

	// Mode describes the deployment mode of the upgrade test, defaults to Cluster
	mode api.DeploymentMode
	// Engine describes the deployment storage engine, defaults to RocksDB
	engine api.StorageEngine

	// fromImage describes the image of the version from which the upgrade should start, defaults to "arangodb/arangodb:<fromVersion>"
	fromImage    string
	fromImageTag string

	// toImage describes the image of the version to which the upgrade should start, defaults to "arangodb/arangodb:<toVersion>"
	toImage    string
	toImageTag string

	toEnterprise   bool
	fromEnterprise bool

	name      string
	shortTest bool
}

type UpgradeTest interface {
	FromVersion() driver.Version
	ToVersion() driver.Version

	Name() string
	FromImage() string
	ToImage() string

	Mode() api.DeploymentMode
	Engine() api.StorageEngine

	IsShortTest() bool
}

func (u *upgradeTest) FromImage() string {
	imageName := "arangodb/arangodb"
	if u.fromEnterprise {
		imageName = "arangodb/enterprise"
	}
	if u.fromImage != "" {
		imageName = u.fromImage
	}
	imageTag := u.fromVersion
	if u.fromImageTag != "" {
		imageTag = u.fromImageTag
	}
	return fmt.Sprintf("%s:%s", imageName, imageTag)
}

func (u *upgradeTest) ToImage() string {
	imageName := "arangodb/arangodb"
	if u.toEnterprise {
		imageName = "arangodb/enterprise"
	}
	if u.toImage != "" {
		imageName = u.toImage
	}
	imageTag := u.toVersion
	if u.toImageTag != "" {
		imageTag = u.toImageTag
	}
	return fmt.Sprintf("%s:%s", imageName, imageTag)
}

func (u *upgradeTest) Mode() api.DeploymentMode {
	if u.mode != "" {
		return u.mode
	}
	return api.DeploymentModeCluster
}

func (u *upgradeTest) Engine() api.StorageEngine {
	if u.engine != "" {
		return u.engine
	}
	return api.StorageEngineRocksDB
}

func (u *upgradeTest) Name() string {
	if u.name != "" {
		return u.name
	}

	return strings.Replace(fmt.Sprintf("%s-to-%s", u.FromVersion(), u.ToVersion()), ".", "-", -1)
}

func (u *upgradeTest) FromVersion() driver.Version {
	return driver.Version(u.fromVersion)
}

func (u *upgradeTest) ToVersion() driver.Version {
	return driver.Version(u.toVersion)
}

func (u *upgradeTest) IsShortTest() bool {
	return u.shortTest
}

func runUpgradeTest(t *testing.T, spec UpgradeTest) {
	if !spec.IsShortTest() {
		longOrSkip(t)
	}

	ns := getNamespace(t)
	kubecli := mustNewKubeClient(t)
	c := kubeArangoClient.MustNewInCluster()

	depl := newDeployment(fmt.Sprintf("tu-%s-%s", spec.Name(), uniuri.NewLen(4)))
	depl.Spec.Mode = api.NewMode(spec.Mode())
	depl.Spec.StorageEngine = api.NewStorageEngine(spec.Engine())
	depl.Spec.TLS = api.TLSSpec{} // should auto-generate cert
	depl.Spec.Image = util.NewString(spec.FromImage())
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
	DBClient := mustNewArangodDatabaseClient(ctx, kubecli, deployment, t, nil)
	if err := waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, spec.FromVersion()); err != nil {
		t.Fatalf("Deployment not healthy in time: %v", err)
	}

	// Try to change image version
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Image = util.NewString(spec.ToImage())
		})
	if err != nil {
		t.Fatalf("Failed to upgrade the Image from version : " + spec.FromImage() + " to version: " + spec.ToImage())
	} else {
		t.Log("Updated deployment")
	}

	if err := waitUntilClusterVersionUp(DBClient, spec.ToVersion()); err != nil {
		t.Errorf("Deployment not healthy in time: %v", err)
	} else {
		t.Log("Deployment healthy")
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
