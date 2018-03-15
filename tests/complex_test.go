package tests

import (
	"context"
	"testing"

	"github.com/dchest/uniuri"
	"k8s.io/client-go/kubernetes"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
)

type kubeSetup struct {
	testName        string
	clientInterface versioned.Interface
	kubecli         kubernetes.Interface
	ns              string
}

type completeDeployment struct {
	instance  *api.ArangoDeployment
	apiobject *api.ArangoDeployment
}

type completeSetup struct {
	kube       kubeSetup
	deployment completeDeployment
	client     driver.Client
}

func prepareKube(t *testing.T, name string) kubeSetup {
	var rv kubeSetup
	rv.clientInterface = client.MustNewInCluster()
	rv.kubecli = mustNewKubeClient(t)
	rv.ns = getNamespace(t)
	rv.testName = name
	return rv
}

func createDeployment(t *testing.T, setup kubeSetup) completeDeployment {
	// Prepare deployment config
	depl := newDeployment("test-" + setup.testName + "-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.DeploymentModeCluster
	depl.Spec.SetDefaults(depl.GetName())

	// Create deployment
	apiObject, err := setup.clientInterface.DatabaseV1alpha().ArangoDeployments(setup.ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(setup.clientInterface, depl.GetName(), setup.ns, deploymentHasState(api.DeploymentStateRunning)); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	return completeDeployment{depl, apiObject}
}

func createClient(t *testing.T, setup kubeSetup, deployment completeDeployment) driver.Client {
	// Create a database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, setup.kubecli, deployment.apiobject, t)

	// Wait for cluster to be completely ready
	if err := waitUntilClusterHealth(client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, deployment.apiobject.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running in expected health in time: %v", err)
	}
	return client
}

func updateCluster(t *testing.T, setup completeSetup) {
	// Add 2 DBServers, 1 coordinator
	updated, err := updateDeployment(setup.kube.clientInterface, setup.deployment.instance.GetName(), setup.kube.ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = 5
		spec.Coordinators.Count = 4
	})

	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(setup.client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-up, in expected health in time: %v", err)
	}

	// Remove 3 DBServers, 2 coordinator
	updated, err = updateDeployment(setup.kube.clientInterface, setup.deployment.instance.GetName(), setup.kube.ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = 3
		spec.Coordinators.Count = 2
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for cluster to reach new size
	if err := waitUntilClusterHealth(setup.client, func(h driver.ClusterHealth) error {
		return clusterHealthEqualsSpec(h, updated.Spec)
	}); err != nil {
		t.Fatalf("Cluster not running, after scale-down, in expected health in time: %v", err)
	}
}

// TestScaleCluster tests scaling up/down the number of DBServers & coordinators
// of a cluster.
func TestComplexCluster(t *testing.T) {
	longOrSkip(t)

	ksetup := prepareKube(t, "complex")
	depl := createDeployment(t, ksetup)
	client := createClient(t, ksetup, depl)

	testSetup := completeSetup{ksetup, depl, client}

	updateCluster(t, testSetup)
	removeDeployment(ksetup.clientInterface, depl.instance.GetName(), ksetup.ns)
}
