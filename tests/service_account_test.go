//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// TestServiceAccountSingle tests the creating of a single server deployment
// with default settings using a custom service account.
func TestServiceAccountSingle(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare service account
	namePrefix := "test-sa-sng-"
	saName := mustCreateServiceAccount(kubecli, namePrefix, ns, t)
	defer deleteServiceAccount(kubecli, saName, ns)

	// Prepare deployment config
	depl := newDeployment(namePrefix + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	depl.Spec.Single.ServiceAccountName = util.NewString(saName)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(context.Background(), depl, metav1.CreateOptions{})
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
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Check service account name
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.Single, saName, t)

	// Check server role
	assert.NoError(t, testServerRole(ctx, client, driver.ServerRoleSingle))
}

// TestServiceAccountActiveFailover tests the creating of a ActiveFailover server deployment
// with default settings using a custom service account.
func TestServiceAccountActiveFailover(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare service account
	namePrefix := "test-sa-rs-"
	saName := mustCreateServiceAccount(kubecli, namePrefix, ns, t)
	defer deleteServiceAccount(kubecli, saName, ns)

	// Prepare deployment config
	depl := newDeployment(namePrefix + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeActiveFailover)
	depl.Spec.Single.ServiceAccountName = util.NewString(saName)
	depl.Spec.Agents.ServiceAccountName = util.NewString(saName)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(context.Background(), depl, metav1.CreateOptions{})
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
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("ActiveFailover servers not running returning version in time: %v", err)
	}

	// Check service account name
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.Single, saName, t)
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.Agents, saName, t)

	// Check server role
	assert.NoError(t, testServerRole(ctx, client, driver.ServerRoleSingleActive))
}

// TestServiceAccountCluster tests the creating of a cluster deployment
// with default settings using a custom service account.
func TestServiceAccountCluster(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare service account
	namePrefix := "test-sa-cls-"
	saName := mustCreateServiceAccount(kubecli, namePrefix, ns, t)
	defer deleteServiceAccount(kubecli, saName, ns)

	// Prepare deployment config
	depl := newDeployment(namePrefix + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Agents.ServiceAccountName = util.NewString(saName)
	depl.Spec.DBServers.ServiceAccountName = util.NewString(saName)
	depl.Spec.Coordinators.ServiceAccountName = util.NewString(saName)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(context.Background(), depl, metav1.CreateOptions{})
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
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Cluster not running returning version in time: %v", err)
	}

	// Check service account name
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.Agents, saName, t)
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.Coordinators, saName, t)
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.DBServers, saName, t)

	// Check server role
	assert.NoError(t, testServerRole(ctx, client, driver.ServerRoleCoordinator))
}

// TestServiceAccountClusterWithSync tests the creating of a cluster deployment
// with default settings and sync enabled using a custom service account.
func TestServiceAccountClusterWithSync(t *testing.T) {
	longOrSkip(t)
	img := getEnterpriseImageOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare service account
	namePrefix := "test-sa-cls-sync-"
	saName := mustCreateServiceAccount(kubecli, namePrefix, ns, t)
	defer deleteServiceAccount(kubecli, saName, ns)

	// Prepare deployment config
	depl := newDeployment(namePrefix + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Image = util.NewString(img)
	depl.Spec.Sync.Enabled = util.NewBool(true)
	depl.Spec.Agents.ServiceAccountName = util.NewString(saName)
	depl.Spec.DBServers.ServiceAccountName = util.NewString(saName)
	depl.Spec.Coordinators.ServiceAccountName = util.NewString(saName)
	depl.Spec.SyncMasters.ServiceAccountName = util.NewString(saName)
	depl.Spec.SyncWorkers.ServiceAccountName = util.NewString(saName)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(context.Background(), depl, metav1.CreateOptions{})
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
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Cluster not running returning version in time: %v", err)
	}

	// Create a syncmaster client
	syncClient := mustNewArangoSyncClient(ctx, kubecli, apiObject, t)

	// Wait for syncmasters to be available
	if err := waitUntilSyncVersionUp(syncClient); err != nil {
		t.Fatalf("SyncMasters not running returning version in time: %v", err)
	}

	// Check service account name
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.Agents, saName, t)
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.Coordinators, saName, t)
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.DBServers, saName, t)
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.SyncMasters, saName, t)
	checkMembersUsingServiceAccount(kubecli, ns, apiObject.Status.Members.SyncWorkers, saName, t)

	// Check server role
	assert.NoError(t, testServerRole(ctx, client, driver.ServerRoleCoordinator))
}

// mustCreateServiceAccount creates an empty service account with random name and returns
// its name. On error, the test is failed.
func mustCreateServiceAccount(kubecli kubernetes.Interface, namePrefix, ns string, t *testing.T) string {
	s := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(namePrefix + uniuri.NewLen(4)),
		},
	}
	if _, err := kubecli.CoreV1().ServiceAccounts(ns).Create(context.Background(), &s, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create service account: %v", err)
	}
	return s.GetName()
}

// deleteServiceAccount deletes a service account with given name in given namespace.
func deleteServiceAccount(kubecli kubernetes.Interface, name, ns string) error {
	if err := kubecli.CoreV1().ServiceAccounts(ns).Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		return maskAny(err)
	}
	return nil
}

// checkMembersUsingServiceAccount checks the serviceAccountName of the pods of all members
// to ensure that is equal to the given serviceAccountName.
func checkMembersUsingServiceAccount(kubecli kubernetes.Interface, ns string, members []api.MemberStatus, serviceAccountName string, t *testing.T) {
	pods := kubecli.CoreV1().Pods(ns)
	for _, m := range members {
		if p, err := pods.Get(context.Background(), m.PodName, metav1.GetOptions{}); err != nil {
			t.Errorf("Failed to get pod for member '%s': %v", m.ID, err)
		} else if p.Spec.ServiceAccountName != serviceAccountName {
			t.Errorf("Expected pod '%s' to have serviceAccountName '%s', got '%s'", p.GetName(), serviceAccountName, p.Spec.ServiceAccountName)
		}
	}
}
