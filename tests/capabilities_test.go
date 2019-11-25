package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddCapabilities(t *testing.T) {
	longOrSkip(t)

	ns := getNamespace(t)
	kubecli := mustNewKubeClient(t)
	c := kubeArangoClient.MustNewInCluster()

	depl := newDeployment(fmt.Sprintf("arangodb-capabilities-test-%s", uniuri.NewLen(4)))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	deployment, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	require.NoErrorf(t, err, "Create deployment failed")
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	deployment, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	require.NoErrorf(t, err, "Deployment not running in time")

	// Create a database client
	ctx := context.Background()
	DBClient := mustNewArangodDatabaseClient(ctx, kubecli, deployment, t, nil)
	err = waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, "")
	require.NoErrorf(t, err, "Deployment not healthy in time")

	expected := []v1.Capability{"SYS_PTRACE", "SYS_CHROOT"}
	deployment, err = updateDeployment(c, depl.GetName(), ns, func(depl *api.DeploymentSpec) {
		depl.Agents.AdditionalCapabilities = expected
	})
	require.NoErrorf(t, err, "Failed to add capabilities metrics")

	var capabilitiesCheck api.ServerGroupFunc = func(group api.ServerGroup, spec api.ServerGroupSpec,
		status *api.MemberStatusList) error {

		if group != api.ServerGroupAgents {
			return nil
		}

		for _, m := range *status {

			pod, err := kubecli.CoreV1().Pods(ns).Get(m.PodName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			server, found := k8sutil.GetContainerByName(pod, k8sutil.ServerContainerName)
			if !found {
				return fmt.Errorf("expected server containter should exist")
			}

			capabilities := server.SecurityContext.Capabilities
			if reconcile.IsAdditionalCapabilitiesChanged(expected, capabilities.Add) {
				return fmt.Errorf("added capabilities have not been changed: expected %v, actual %v",
					expected, server.SecurityContext.Capabilities.Add)
			}

			if reconcile.IsAdditionalCapabilitiesChanged([]v1.Capability{"ALL"}, capabilities.Drop) {
				return fmt.Errorf("dropped capabilities should equal to 'ALL'. actual %v",
					server.SecurityContext.Capabilities.Drop)
			}
		}
		return nil
	}
	_, err = waitUntilDeploymentMembers(c, deployment.GetName(), ns, capabilitiesCheck, 5*time.Minute)
	require.NoError(t, err)

	err = waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, "")
	require.NoErrorf(t, err, "Deployment not healthy in time")

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
