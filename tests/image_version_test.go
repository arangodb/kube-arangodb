package tests

import (
	"context"
	"fmt"
	"time"
	"os"
	"testing"

	"github.com/dchest/uniuri"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
)


// TestImmutableStorageEngine
// Tests that storage engine of deployed cluster cannot be changed
func TestDowngradePatchLevel(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-dpl-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.DeploymentModeCluster
	depl.Spec.SetDefaults(depl.GetName())
	depl.Spec.Image = "arangodb/arangodb:3.3.3"

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, 
		deploymentHasState(api.DeploymentStateRunning)); err != nil {
			t.Fatalf("Deployment not running in time: %v", err)
		}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Deployed version: %s\n", depl.Spec.Image)

	// Try to upgrade to image version to 3.3.4 ==================
	updated, err := updateDeployment(c, depl.GetName(), ns,
		func(spec *api.DeploymentSpec) {
			spec.Image = "arangodb/arangodb:3.3.4"
		}); 
	if err != nil {
		t.Fatalf("Failed to upgrade the Image to 3.3.4")
	} 

	time.Sleep(10 * time.Second)

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-up, in expected health in time: %v", err)
	}

	// Try to downgrade to image version to 3.3.2 ==================
	updated, err = updateDeployment(c, depl.GetName(), ns,
		func(spec *api.DeploymentSpec) {
			spec.Image = "arangodb/arangodb:3.3.4"
		}); 
	if err != nil {
		t.Fatalf("Failed to upgrade the Image to 3.3.4")
	} 

	time.Sleep(10 * time.Second)

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-up, in expected health in time: %v", err)
	}

	// Try to downgrade to image version to 3.2.12 =================
	updated, err = updateDeployment(c, depl.GetName(), ns,
		func(spec *api.DeploymentSpec) {
			spec.Image = "arangodb/arangodb:3.2.12"
		}); 
	if err != nil {
		t.Fatalf("Failed to update the image setting: %v", err)
	} 

	// Wait for StorageEngine parameter to be back to RocksDB
	if _, err := waitUntilDeployment(c, depl.GetName(), ns,
		func(depl *api.ArangoDeployment) error {
			if depl.Spec.StorageEngine == "arangodb/arangodb:3.3.3" {
				return nil
			} 
			return fmt.Errorf("One must not downgrade to a lower minor revision")
		}); err != nil {
			t.Fatalf("Deployment was downgraded to lower minor version 3.2.12")
		}
		
		// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
