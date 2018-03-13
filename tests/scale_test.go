package tests

import (
	"context"
	"testing"

	"github.com/dchest/uniuri"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
)

// TestScaleCluster tests scaling up/down the number of DBServers & coordinators
// of a cluster.
func TestScaleCluster(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-scale-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.DeploymentModeCluster
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

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Add 2 DBServers, 1 coordinator
	updated, err := updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = 5
		spec.Coordinators.Count = 4
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-up, in expected health in time: %v", err)
	}

	// Remove 3 DBServers, 2 coordinator
	updated, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = 3
		spec.Coordinators.Count = 2
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-down, in expected health in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
