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

// TestResiliencePod
// Tests handling of individual pod deletions
func TestResiliencePod(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// Prepare deployment config
	depl := newDeployment("test-pod-resilience" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentHasState(api.DeploymentStateRunning)); err != nil {
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

// TestResiliencePVC
// Tests handling of individual pod deletions
func TestResiliencePVC(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-pvc-resilience" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentHasState(api.DeploymentStateRunning)); err != nil {
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

	// Delete one pvc after the other
	apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		if group == api.ServerGroupCoordinators {
			// Coordinators have no PVC
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
					return maskAny(err)
				}
				if pvc.GetUID() == originalPVC.GetUID() {
					return fmt.Errorf("Still original pvc")
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
	depl := newDeployment("test-service-resilience" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	apiObject, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentHasState(api.DeploymentStateRunning)); err != nil {
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
