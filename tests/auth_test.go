package tests

import (
	"context"
	"testing"
	"time"

	"github.com/dchest/uniuri"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/client"
	"github.com/arangodb/k8s-operator/pkg/util/arangod"
)

// TestAuthenticationSingleDefaultSecret creating a single server
// with default authentication (on) using a generated JWT secret.
func TestAuthenticationSingleDefaultSecret(t *testing.T) {
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-auth-sng-def-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.DeploymentModeSingle
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

	// Secret must now exist
	if _, err := waitUntilSecret(kubecli, depl.Spec.Authentication.JWTSecretName, ns, nil, time.Second); err != nil {
		t.Fatalf("JWT secret '%s' not found: %v", depl.Spec.Authentication.JWTSecretName, err)
	}

	// Create a database client
	ctx := arangod.WithRequireAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)

	// Secret must no longer exist
	if err := waitUntilSecretNotFound(kubecli, depl.Spec.Authentication.JWTSecretName, ns, time.Minute); err != nil {
		t.Fatalf("JWT secret '%s' still found: %v", depl.Spec.Authentication.JWTSecretName, err)
	}
}

// TestAuthenticationNoneSingle creating a single server
// with authentication set to `None`.
func TestAuthenticationNoneSingle(t *testing.T) {
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-auth-none-sng-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.DeploymentModeSingle
	depl.Spec.Authentication.JWTSecretName = api.JWTSecretNameDisabled
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
	ctx := arangod.WithSkipAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
