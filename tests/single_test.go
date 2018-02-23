package tests

import (
	"testing"

	"github.com/dchest/uniuri"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/client"
)

// TestSimpleSingle tests the creating of a single server deployment
// with default settings.
func TestSimpleSingle(t *testing.T) {
	c := client.MustNewInCluster()
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-single-" + uniuri.NewLen(4))

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentHasState(api.DeploymentStateRunning)); err != nil {
		t.Errorf("Deployment not running in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
