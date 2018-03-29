package tests

import (
	"context"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
)

// TestSimpleSingle tests the creating of a single server deployment
// with default settings.
func TestSimpleSingle(t *testing.T) {
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-sng-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Check server role
	assert.NoError(t, client.SynchronizeEndpoints(ctx))
	role, err := client.ServerRole(ctx)
	assert.NoError(t, err)
	assert.Equal(t, driver.ServerRoleSingle, role)
}

// TestSimpleResilientSingle tests the creating of a resilientsingle server deployment
// with default settings.
func TestSimpleResilientSingle(t *testing.T) {
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-rs-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeResilientSingle)

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("ResilientSingle servers not running returning version in time: %v", err)
	}

	// Check server role
	assert.NoError(t, client.SynchronizeEndpoints(ctx))
	role, err := client.ServerRole(ctx)
	assert.NoError(t, err)
	assert.Equal(t, driver.ServerRoleSingleActive, role)
}

// TestSimpleCluster tests the creating of a cluster deployment
// with default settings.
func TestSimpleCluster(t *testing.T) {
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-cls-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Cluster not running returning version in time: %v", err)
	}

	// Check server role
	assert.NoError(t, client.SynchronizeEndpoints(ctx))
	role, err := client.ServerRole(ctx)
	assert.NoError(t, err)
	assert.Equal(t, driver.ServerRoleCoordinator, role)
}
