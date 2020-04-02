package tests

import (
	"fmt"
	"testing"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func isPDBAsExpected(kube kubernetes.Interface, name, ns string, expectedMinAvailable int) error {
	pdb, err := kube.PolicyV1beta1().PodDisruptionBudgets(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if pdb.Spec.MinAvailable.IntValue() != expectedMinAvailable {
		return fmt.Errorf("PDB %s does not have expected minAvailable, found: %d, expected: %d", name, pdb.Spec.MinAvailable.IntValue(), expectedMinAvailable)
	}
	return nil
}

func waitForPDBAsExpected(kube kubernetes.Interface, name, ns string, expectedMinAvailable int) error {
	return retry.Retry(func() error {
		return isPDBAsExpected(kube, name, ns, expectedMinAvailable)
	}, 20*time.Second)
}

func waitForPDBsOfDeployment(kube kubernetes.Interface, apiObject *api.ArangoDeployment) error {
	spec := apiObject.Spec
	return retry.Retry(func() error {
		if spec.Mode.HasAgents() {
			if err := isPDBAsExpected(kube, resources.PDBNameForGroup(apiObject.GetName(), api.ServerGroupAgents), apiObject.GetNamespace(), spec.GetServerGroupSpec(api.ServerGroupAgents).GetCount()-1); err != nil {
				return err
			}
		}
		if spec.Mode.HasCoordinators() {
			if err := isPDBAsExpected(kube, resources.PDBNameForGroup(apiObject.GetName(), api.ServerGroupCoordinators), apiObject.GetNamespace(),
				min(spec.GetServerGroupSpec(api.ServerGroupCoordinators).GetCount()-1, 2)); err != nil {
				return err
			}
		}
		if spec.Mode.HasDBServers() {
			if err := isPDBAsExpected(kube, resources.PDBNameForGroup(apiObject.GetName(), api.ServerGroupDBServers), apiObject.GetNamespace(), spec.GetServerGroupSpec(api.ServerGroupDBServers).GetCount()-1); err != nil {
				return err
			}
		}
		return nil
	}, 20*time.Second)
}

// TestPDBCreate create a deployment and check if the PDBs are created. Then rescale the cluster and check if the PDBs are
// modified accordingly.
func TestPDBCreate(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-pdb-create-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Environment = api.NewEnvironment(api.EnvironmentProduction)
	depl.Spec.TLS = api.TLSSpec{CASecretName: util.NewString("None")}
	depl.Spec.DBServers.Count = util.NewInt(2)
	depl.Spec.Coordinators.Count = util.NewInt(2)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// This test failes to validate the spec if no image is set explicitly because this is required in production mode
	if depl.Spec.Image == nil {
		depl.Spec.Image = util.NewString("arangodb/arangodb:latest")
	}
	assert.NoError(t, depl.Spec.Validate())

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	if err := waitForPDBsOfDeployment(kubecli, apiObject); err != nil {
		t.Errorf("PDBs not as expected: %v", err)
	}

	apiObject, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.DBServers.Count = util.NewInt(3)
		spec.Coordinators.Count = util.NewInt(3)
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Wait for deployment to be ready
	if _, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady()); err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}
	// Check if the PDBs have grown too
	if err := waitForPDBsOfDeployment(kubecli, apiObject); err != nil {
		t.Errorf("PDBs not as expected: %v", err)
	}

	removeDeployment(c, depl.GetName(), ns)
}
