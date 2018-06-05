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
// Author Ewout Prangsma
//

package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dchest/uniuri"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// TestAuthenticationSingleDefaultSecret creating a single server
// with default authentication (on) using a generated JWT secret.
func TestAuthenticationSingleDefaultSecret(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-auth-sng-def-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	depl.Spec.SetDefaults(depl.GetName())

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Secret must now exist
	if _, err := waitUntilSecret(kubecli, depl.Spec.Authentication.GetJWTSecretName(), ns, nil, time.Second); err != nil {
		t.Fatalf("JWT secret '%s' not found: %v", depl.Spec.Authentication.GetJWTSecretName(), err)
	}

	// Create a database client
	ctx := arangod.WithRequireAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)

	// Secret must no longer exist
	if err := waitUntilSecretNotFound(kubecli, depl.Spec.Authentication.GetJWTSecretName(), ns, time.Minute); err != nil {
		t.Fatalf("JWT secret '%s' still found: %v", depl.Spec.Authentication.GetJWTSecretName(), err)
	}
}

// TestAuthenticationSingleCustomSecret creating a single server
// with default authentication (on) using a user created JWT secret.
func TestAuthenticationSingleCustomSecret(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-auth-sng-cst-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	depl.Spec.Authentication.JWTSecretName = util.NewString(strings.ToLower(uniuri.New()))
	depl.Spec.SetDefaults(depl.GetName())

	// Create secret
	if err := k8sutil.CreateTokenSecret(kubecli.CoreV1(), depl.Spec.Authentication.GetJWTSecretName(), ns, "foo", nil); err != nil {
		t.Fatalf("Create JWT secret failed: %v", err)
	}

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := arangod.WithRequireAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)

	// Secret must still exist
	if _, err := waitUntilSecret(kubecli, depl.Spec.Authentication.GetJWTSecretName(), ns, nil, time.Second); err != nil {
		t.Fatalf("JWT secret '%s' not found: %v", depl.Spec.Authentication.GetJWTSecretName(), err)
	}

	// Cleanup secret
	removeSecret(kubecli, depl.Spec.Authentication.GetJWTSecretName(), ns)
}

// TestAuthenticationNoneSingle creating a single server
// with authentication set to `None`.
func TestAuthenticationNoneSingle(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-auth-none-sng-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	depl.Spec.Authentication.JWTSecretName = util.NewString(api.JWTSecretNameDisabled)
	depl.Spec.SetDefaults(depl.GetName())

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := arangod.WithSkipAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}

// TestAuthenticationClusterDefaultSecret creating a cluster
// with default authentication (on) using a generated JWT secret.
func TestAuthenticationClusterDefaultSecret(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-auth-cls-def-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName())

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Secret must now exist
	if _, err := waitUntilSecret(kubecli, depl.Spec.Authentication.GetJWTSecretName(), ns, nil, time.Second); err != nil {
		t.Fatalf("JWT secret '%s' not found: %v", depl.Spec.Authentication.GetJWTSecretName(), err)
	}

	// Create a database client
	ctx := arangod.WithRequireAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)

	// Secret must no longer exist
	if err := waitUntilSecretNotFound(kubecli, depl.Spec.Authentication.GetJWTSecretName(), ns, time.Minute); err != nil {
		t.Fatalf("JWT secret '%s' still found: %v", depl.Spec.Authentication.GetJWTSecretName(), err)
	}
}

// TestAuthenticationClusterCustomSecret creating a cluster
// with default authentication (on) using a user created JWT secret.
func TestAuthenticationClusterCustomSecret(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-auth-cls-cst-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Authentication.JWTSecretName = util.NewString(strings.ToLower(uniuri.New()))
	depl.Spec.SetDefaults(depl.GetName())

	// Create secret
	if err := k8sutil.CreateTokenSecret(kubecli.CoreV1(), depl.Spec.Authentication.GetJWTSecretName(), ns, "foo", nil); err != nil {
		t.Fatalf("Create JWT secret failed: %v", err)
	}

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := arangod.WithRequireAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)

	// Secret must still exist
	if _, err := waitUntilSecret(kubecli, depl.Spec.Authentication.GetJWTSecretName(), ns, nil, time.Second); err != nil {
		t.Fatalf("JWT secret '%s' not found: %v", depl.Spec.Authentication.GetJWTSecretName(), err)
	}

	// Cleanup secret
	removeSecret(kubecli, depl.Spec.Authentication.GetJWTSecretName(), ns)
}

// TestAuthenticationNoneCluster creating a cluster
// with authentication set to `None`.
func TestAuthenticationNoneCluster(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-auth-none-cls-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Authentication.JWTSecretName = util.NewString(api.JWTSecretNameDisabled)
	depl.Spec.SetDefaults(depl.GetName())

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := arangod.WithSkipAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
