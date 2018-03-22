package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dchest/uniuri"	
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	
	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"

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
	depl.Spec.Mode = api.DeploymentModeCluster
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

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
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Delete one pod after the other			
	pods, err := kubecli.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Could not find any pods in the %s, namespace: %v\n", ns, err)
	}
	fmt.Fprintf(os.Stderr, 
		"There are %d pods in the %s namespace\n", len(pods.Items), ns)
	for _, pod := range pods.Items {
		if pod.GetName() == "arangodb-operator-test" { continue }
		fmt.Fprintf(os.Stderr, 
			"Deleting pod %s in the %s namespace\n", pod.GetName(), ns)
		kubecli.CoreV1().Pods(ns).Delete(pod.GetName(),&metav1.DeleteOptions{})
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

// TestResilienceService
// Tests handling of individual service deletions
func TestResilienceService(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-service-resilience" + uniuri.NewLen(4))
	depl.Spec.Mode = api.DeploymentModeCluster
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

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
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, apiObject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}

	// Delete one pod after the other			
	services, err := kubecli.CoreV1().Services(ns).List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Could not find any services in the %s, namespace: %v\n", ns, err)
	}
	fmt.Fprintf(os.Stderr, "There are %d pods in the %s namespace \n", len(services.Items), ns)
	for _, service := range services.Items {
		kubecli.CoreV1().Services(ns).Delete(service.GetName(),&metav1.DeleteOptions{})
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
