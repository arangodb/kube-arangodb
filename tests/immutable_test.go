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
// Author Kaveh Vahedipour
// Author Ewout Prangsma
//

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dchest/uniuri"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
)

// TestImmutableFields tests that several immutable fields in the deployment
// spec are reverted to their original value.
func TestImmutableFields(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)
	revertTimeout := time.Second * 30

	// Prepare deployment config
	depl := newDeployment("test-ise-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	depl.Spec.SetDefaults(depl.GetName())

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentHasState(api.DeploymentStateRunning)); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server to be completely ready
	if err := waitUntilVersionUp(client); err != nil {
		t.Fatalf("Single server not up in time: %v", err)
	}

	// Try to reset storageEngine ===============================================
	if _, err := updateDeployment(c, depl.GetName(), ns,
		func(spec *api.DeploymentSpec) {
			spec.StorageEngine = api.NewStorageEngine(api.StorageEngineMMFiles)
		}); err != nil {
		t.Fatalf("Failed to update the StorageEngine setting: %v", err)
	}

	// Wait for StorageEngine parameter to be back to RocksDB
	if _, err := waitUntilDeployment(c, depl.GetName(), ns,
		func(depl *api.ArangoDeployment) error {
			if api.StorageEngineOrDefault(depl.Spec.StorageEngine) == api.StorageEngineRocksDB {
				return nil
			}
			return fmt.Errorf("StorageEngine not back to %s", api.StorageEngineRocksDB)
		}, revertTimeout); err != nil {
		t.Errorf("StorageEngine parameter is immutable: %v", err)
	}

	/*
		 Secrets are a special case that we'll deal with later
		// Try to reset the RocksDB encryption key ==================================
		if _, err := updateDeployment(c, depl.GetName(), ns,
			func(spec *api.DeploymentSpec) {
				spec.RocksDB.Encryption.KeySecretName = util.NewString("foobarbaz")
			}); err != nil {
			t.Fatalf("Failed to update the RocksDB encryption key: %v", err)
		}

		// Wait for deployment mode to be set back to cluster
		if _, err := waitUntilDeployment(c, depl.GetName(), ns,
			func(depl *api.ArangoDeployment) error {
				if util.StringOrDefault(depl.Spec.RocksDB.Encryption.KeySecretName) == "test.encryption.keySecretName" {
					return nil
				}
				return fmt.Errorf("RocksDB encryption key not back to %s", "test.encryption.keySecretName")
			}, revertTimeout); err != nil {
			t.Errorf("RocksDB encryption key is mutable: %v", err)
		}
	*/

	// Try to reset the deployment type ==========================================
	if _, err := updateDeployment(c, depl.GetName(), ns,
		func(spec *api.DeploymentSpec) {
			spec.Mode = api.NewMode(api.DeploymentModeCluster)
		}); err != nil {
		t.Fatalf("Failed to update the deployment mode: %v", err)
	}

	// Wait for deployment mode to be set back to cluster
	if _, err := waitUntilDeployment(c, depl.GetName(), ns,
		func(depl *api.ArangoDeployment) error {
			expected := api.DeploymentModeSingle
			if api.ModeOrDefault(depl.Spec.Mode) == expected {
				return nil
			}
			return fmt.Errorf("Deployment mode not back to %s", expected)
		}, revertTimeout); err != nil {
		t.Errorf("Deployment mode is mutable: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
