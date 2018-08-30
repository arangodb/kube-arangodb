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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dchest/uniuri"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
)

// TestResiliencePod
// Tests handling of individual pod deletions
func TestResiliencePod(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// Prepare deployment config
	depl := newDeployment("test-pod-resilience-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Fetch latest status so we know all member details
	apiObject, err = c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	// Delete one pod after the other
	apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for _, m := range *status {
			// Get current pod so we can compare UID later
			originalPod, err := kubecli.CoreV1().Pods(ns).Get(m.PodName, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("Failed to get pod %s: %v", m.PodName, err)
			}
			if err := kubecli.CoreV1().Pods(ns).Delete(m.PodName, &metav1.DeleteOptions{}); err != nil {
				t.Fatalf("Failed to delete pod %s: %v", m.PodName, err)
			}
			// Wait for pod to return with different UID
			op := func() error {
				pod, err := kubecli.CoreV1().Pods(ns).Get(m.PodName, metav1.GetOptions{})
				if err != nil {
					return maskAny(err)
				}
				if pod.GetUID() == originalPod.GetUID() {
					return fmt.Errorf("Still original pod")
				}
				return nil
			}
			if err := retry.Retry(op, time.Minute); err != nil {
				t.Fatalf("Pod did not restart: %v", err)
			}
			// Wait for deployment to be ready
			if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
				t.Fatalf("Deployment not running in time: %v", err)
			}
			// Wait for cluster to be completely ready
			if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
				return clusterHealthEqualsSpec(h, apiObject.Spec)
			}); err != nil {
				t.Fatalf("Cluster not running in expected health in time: %v", err)
			}
		}
		return nil
	}, &apiObject.Status)

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}

// TestResiliencePVCAgents
// Tests handling of individual PVCs of agents being deleted
func TestResiliencePVCAgents(t *testing.T) {
	testResiliencePVC(api.ServerGroupAgents, t)
}

// TestResiliencePVCDBServers
// Tests handling of individual PVCs of dbservers being deleted
func TestResiliencePVCDBServers(t *testing.T) {
	testResiliencePVC(api.ServerGroupDBServers, t)
}

// testResiliencePVC
// Tests handling of individual PVCs of given group being deleted
func testResiliencePVC(testGroup api.ServerGroup, t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment(fmt.Sprintf("test-pvc-resilience-%s-%s", testGroup.AsRoleAbbreviated(), uniuri.NewLen(4)))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Fetch latest status so we know all member details
	apiObject, err = c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	// Delete one pvc after the other
	apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		if group != testGroup {
			// We only test a specific group here
			return nil
		}
		for _, m := range *status {
			// Get current pvc so we can compare UID later
			originalPVC, err := kubecli.CoreV1().PersistentVolumeClaims(ns).Get(m.PersistentVolumeClaimName, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("Failed to get pvc %s: %v", m.PersistentVolumeClaimName, err)
			}
			if err := kubecli.CoreV1().PersistentVolumeClaims(ns).Delete(m.PersistentVolumeClaimName, &metav1.DeleteOptions{}); err != nil {
				t.Fatalf("Failed to delete pvc %s: %v", m.PersistentVolumeClaimName, err)
			}
			// Wait for pvc to return with different UID
			op := func() error {
				pvc, err := kubecli.CoreV1().PersistentVolumeClaims(ns).Get(m.PersistentVolumeClaimName, metav1.GetOptions{})
				if err != nil {
					if k8sutil.IsNotFound(err) && group == api.ServerGroupDBServers {
						// DBServer member is completely replaced when cleaned out, so the PVC will have a different name also
						return nil
					}
					return maskAny(err)
				}
				if pvc.GetUID() == originalPVC.GetUID() {
					return fmt.Errorf("Still original pvc")
				}
				return nil
			}
			if err := retry.Retry(op, time.Minute*2); err != nil {
				t.Fatalf("PVC did not restart: %v", err)
			}
			// Wait for deployment to be ready
			if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
				t.Fatalf("Deployment not running in time: %v", err)
			}
			// Wait for cluster to be completely ready
			if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
				return clusterHealthEqualsSpec(h, apiObject.Spec)
			}); err != nil {
				t.Fatalf("Cluster not running in expected health in time: %v", err)
			}
		}
		return nil
	}, &apiObject.Status)

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}

// TestResiliencePVDBServer
// Tests handling of entire PVs of dbservers being removed.
func TestResiliencePVDBServer(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-pv-prmr-resi-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Fetch latest status so we know all member details
	apiObject, err = c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	// Delete one pv, pvc & pod after the other
	apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		if group != api.ServerGroupDBServers {
			// Agents cannot be replaced with a new ID
			// Coordinators, Sync masters/workers have no persistent storage
			return nil
		}
		for i, m := range *status {
			// Only test first 2
			if i >= 2 {
				continue
			}
			// Get current pvc so we can compare UID later
			originalPVC, err := kubecli.CoreV1().PersistentVolumeClaims(ns).Get(m.PersistentVolumeClaimName, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("Failed to get pvc %s: %v", m.PersistentVolumeClaimName, err)
			}
			// Get current pv
			pvName := originalPVC.Spec.VolumeName
			require.NotEmpty(t, pvName, "VolumeName of %s must be non-empty", originalPVC.GetName())
			// Delete PV
			if err := kubecli.CoreV1().PersistentVolumes().Delete(pvName, &metav1.DeleteOptions{}); err != nil {
				t.Fatalf("Failed to delete pv %s: %v", pvName, err)
			}
			// Delete PVC
			if err := kubecli.CoreV1().PersistentVolumeClaims(ns).Delete(m.PersistentVolumeClaimName, &metav1.DeleteOptions{}); err != nil {
				t.Fatalf("Failed to delete pvc %s: %v", m.PersistentVolumeClaimName, err)
			}
			// Delete Pod
			/*if err := kubecli.CoreV1().Pods(ns).Delete(m.PodName, &metav1.DeleteOptions{}); err != nil {
				t.Fatalf("Failed to delete pod %s: %v", m.PodName, err)
			}*/
			// Wait for cluster to be healthy again with the same number of
			// dbservers, but the current dbserver being replaced.
			expectedDBServerCount := apiObject.Spec.DBServers.GetCount()
			unexpectedID := m.ID
			pred := func(depl *api.ArangoDeployment) error {
				if len(depl.Status.Members.DBServers) != expectedDBServerCount {
					return maskAny(fmt.Errorf("Expected %d dbservers, got %d", expectedDBServerCount, len(depl.Status.Members.DBServers)))
				}
				if depl.Status.Members.ContainsID(unexpectedID) {
					return maskAny(fmt.Errorf("Member %s should be gone", unexpectedID))
				}
				return nil
			}
			if _, err := waitUntilDeployment(c, apiObject.GetName(), ns, pred, time.Minute*5); err != nil {
				t.Fatalf("Deployment not ready in time: %v", err)
			}
			// Wait for cluster to be completely ready
			if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
				return clusterHealthEqualsSpec(h, apiObject.Spec)
			}); err != nil {
				t.Fatalf("Cluster not running in expected health in time: %v", err)
			}
		}
		return nil
	}, &apiObject.Status)

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}

// TestResilienceService
// Tests handling of individual service deletions
func TestResilienceService(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-service-resilience-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Fetch latest status so we know all member details
	apiObject, err = c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	// Delete database service
	// Get current pod so we can compare UID later
	serviceName := apiObject.Status.ServiceName
	originalService, err := kubecli.CoreV1().Services(ns).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get service %s: %v", serviceName, err)
	}
	if err := kubecli.CoreV1().Services(ns).Delete(serviceName, &metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Failed to delete service %s: %v", serviceName, err)
	}
	// Wait for service to return with different UID
	op := func() error {
		service, err := kubecli.CoreV1().Services(ns).Get(serviceName, metav1.GetOptions{})
		if err != nil {
			return maskAny(err)
		}
		if service.GetUID() == originalService.GetUID() {
			return fmt.Errorf("Still original service")
		}
		return nil
	}
	if err := retry.Retry(op, time.Minute); err != nil {
		t.Fatalf("PVC did not restart: %v", err)
	}
	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
