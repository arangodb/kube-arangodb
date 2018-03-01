package tests

import (
	"context"
	"testing"

	"github.com/dchest/uniuri"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/client"
)

// TestSimpleSingle tests the creating of a single server deployment
// with default settings.
func TestSimpleSingle(t *testing.T) {
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-single-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.DeploymentModeSingle

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentHasState(api.DeploymentStateRunning))
	if err != nil {
		t.Errorf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}
}
