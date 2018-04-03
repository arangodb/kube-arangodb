package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dchest/uniuri"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
)

// TestMemberResilienceAgents creates a cluster and removes a
// specific agent pod 5 times. Each time it waits for a new pod to arrive.
// After 5 times, the member should be replaced by another member with the same ID.
func TestMemberResilienceAgents(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-member-res-agnt-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
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

	// Fetch latest status so we know all member details
	apiObject, err = c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	// Pick an agent to be deleted 5 times
	targetAgent := apiObject.Status.Members.Agents[0]
	for i := 0; i < 5; i++ {
		// Get current pod so we can compare UID later
		originalPod, err := kubecli.CoreV1().Pods(ns).Get(targetAgent.PodName, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get pod %s: %v", targetAgent.PodName, err)
		}
		if err := kubecli.CoreV1().Pods(ns).Delete(targetAgent.PodName, &metav1.DeleteOptions{}); err != nil {
			t.Fatalf("Failed to delete pod %s: %v", targetAgent.PodName, err)
		}
		if i < 4 {
			// Wait for pod to return with different UID
			op := func() error {
				pod, err := kubecli.CoreV1().Pods(ns).Get(targetAgent.PodName, metav1.GetOptions{})
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
		} else {
			// Wait for member to be replaced
			op := func() error {
				updatedObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
				if err != nil {
					return maskAny(err)
				}
				m, _, found := updatedObject.Status.Members.ElementByID(targetAgent.ID)
				if !found {
					return maskAny(fmt.Errorf("Member %s not found", targetAgent.ID))
				}
				if m.CreatedAt.Equal(&targetAgent.CreatedAt) {
					return maskAny(fmt.Errorf("Member %s still not replaced", targetAgent.ID))
				}
				return nil
			}
			if err := retry.Retry(op, time.Minute); err != nil {
				t.Fatalf("Member failure did not succeed: %v", err)
			}
		}
		// Wait for cluster to be completely ready
		if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
			return clusterHealthEqualsSpec(h, apiObject.Spec)
		}); err != nil {
			t.Fatalf("Cluster not running in expected health in time: %v", err)
		}
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}

// TestMemberResilienceCoordinators creates a cluster and removes a
// specific coordinator pod 5 times. Each time it waits for a new pod to arrive.
// After 5 times, the member should be replaced by another member.
func TestMemberResilienceCoordinators(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-member-res-crdn-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
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

	// Fetch latest status so we know all member details
	apiObject, err = c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	// Pick a coordinator to be deleted 5 times
	targetCoordinator := apiObject.Status.Members.Coordinators[0]
	for i := 0; i < 5; i++ {
		// Get current pod so we can compare UID later
		originalPod, err := kubecli.CoreV1().Pods(ns).Get(targetCoordinator.PodName, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get pod %s: %v", targetCoordinator.PodName, err)
		}
		if err := kubecli.CoreV1().Pods(ns).Delete(targetCoordinator.PodName, &metav1.DeleteOptions{}); err != nil {
			t.Fatalf("Failed to delete pod %s: %v", targetCoordinator.PodName, err)
		}
		if i < 4 {
			// Wait for pod to return with different UID
			op := func() error {
				pod, err := kubecli.CoreV1().Pods(ns).Get(targetCoordinator.PodName, metav1.GetOptions{})
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
		} else {
			// Wait for member to be replaced
			op := func() error {
				updatedObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
				if err != nil {
					return maskAny(err)
				}
				if updatedObject.Status.Members.ContainsID(targetCoordinator.ID) {
					return maskAny(fmt.Errorf("Member %s still not replaced", targetCoordinator.ID))
				}
				return nil
			}
			if err := retry.Retry(op, time.Minute); err != nil {
				t.Fatalf("Member failure did not succeed: %v", err)
			}
		}
		// Wait for cluster to be completely ready
		if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
			return clusterHealthEqualsSpec(h, apiObject.Spec)
		}); err != nil {
			t.Fatalf("Cluster not running in expected health in time: %v", err)
		}
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}

// TestMemberResilienceDBServers creates a cluster and removes a
// specific dbserver pod 5 times. Each time it waits for a new pod to arrive.
// After 5 times, the member should be replaced by another member.
func TestMemberResilienceDBServers(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-member-res-prmr-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
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

	// Fetch latest status so we know all member details
	apiObject, err = c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get deployment: %v", err)
	}

	// Pick a coordinator to be deleted 5 times
	targetServer := apiObject.Status.Members.DBServers[0]
	for i := 0; i < 5; i++ {
		// Get current pod so we can compare UID later
		originalPod, err := kubecli.CoreV1().Pods(ns).Get(targetServer.PodName, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get pod %s: %v", targetServer.PodName, err)
		}
		if err := kubecli.CoreV1().Pods(ns).Delete(targetServer.PodName, &metav1.DeleteOptions{}); err != nil {
			t.Fatalf("Failed to delete pod %s: %v", targetServer.PodName, err)
		}
		if i < 4 {
			// Wait for pod to return with different UID
			op := func() error {
				pod, err := kubecli.CoreV1().Pods(ns).Get(targetServer.PodName, metav1.GetOptions{})
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
		} else {
			// Wait for member to be replaced
			op := func() error {
				updatedObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
				if err != nil {
					return maskAny(err)
				}
				if updatedObject.Status.Members.ContainsID(targetServer.ID) {
					return maskAny(fmt.Errorf("Member %s still not replaced", targetServer.ID))
				}
				return nil
			}
			if err := retry.Retry(op, time.Minute); err != nil {
				t.Fatalf("Member failure did not succeed: %v", err)
			}
		}
		// Wait for cluster to be completely ready
		if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
			return clusterHealthEqualsSpec(h, apiObject.Spec)
		}); err != nil {
			t.Fatalf("Cluster not running in expected health in time: %v", err)
		}
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
