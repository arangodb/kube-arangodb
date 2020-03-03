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
	"fmt"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// TestChangeArgsAgents tests the creating of an active failover deployment
// with default settings and once ready changes the arguments of the agents.
func TestChangeArgsAgents(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-chga-rs-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeActiveFailover)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
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

	// Check server role
	assert.NoError(t, testServerRole(ctx, client, driver.ServerRoleSingleActive))

	// Now change agent arguments
	if _, err := updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.Agents.Args = []string{"--log.level=DEBUG"}
	}); err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait until all agents have the right arguments
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, func(d *api.ArangoDeployment) error {
		members := d.Status.Members
		if len(members.Agents) != 3 {
			return fmt.Errorf("Expected 3 agents, got %d", len(members.Agents))
		}
		pods := kubecli.CoreV1().Pods(ns)
		for _, m := range members.Agents {
			pod, err := pods.Get(m.PodName, metav1.GetOptions{})
			if err != nil {
				return maskAny(err)
			}
			found := false
			for _, c := range pod.Spec.Containers {
				if c.Name != k8sutil.ServerContainerName {
					continue
				}
				// Check command
				for _, a := range append(c.Args, c.Command...) {
					if a == "--log.level=DEBUG" {
						found = true
					}
				}
			}
			if !found {
				return fmt.Errorf("Did not find new argument")
			}
		}
		return nil
	}, time.Minute*10); err != nil {
		t.Fatalf("Deployment not updated in time: %v", err)
	}
}

// TestChangeArgsDBServer tests the creating of a cluster deployment
// with default settings and once ready changes the arguments of the dbservers.
func TestChangeArgsDBServer(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-chga-db-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
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

	// Wait for cluster available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Cluster servers not running returning version in time: %v", err)
	}

	// Now change agent arguments
	if _, err := updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Args = []string{"--log.level=DEBUG"}
	}); err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait until all dbservers have the right arguments
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, func(d *api.ArangoDeployment) error {
		members := d.Status.Members
		if len(members.DBServers) != 3 {
			return fmt.Errorf("Expected 3 dbservers, got %d", len(members.DBServers))
		}
		pods := kubecli.CoreV1().Pods(ns)
		for _, m := range members.DBServers {
			pod, err := pods.Get(m.PodName, metav1.GetOptions{})
			if err != nil {
				return maskAny(err)
			}
			found := false
			for _, c := range pod.Spec.Containers {
				if c.Name != k8sutil.ServerContainerName {
					continue
				}
				// Check command
				for _, a := range append(c.Args, c.Command...) {
					if a == "--log.level=DEBUG" {
						found = true
					}
				}
			}
			if !found {
				return fmt.Errorf("Did not find new argument")
			}
		}
		return nil
	}, time.Minute*10); err != nil {
		t.Fatalf("Deployment not updated in time: %v", err)
	}
}
