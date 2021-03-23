package tests

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/dchest/uniuri"
)

func TestIsVersionSet(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	expectedVersion := driver.Version("3.3.17")
	// Prepare deployment config
	depl := newDeployment("test-auth-sng-def-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	depl.Spec.SetDefaults(depl.GetName())
	depl.Spec.Image = util.NewString("arangodb/arangodb:" + string(expectedVersion))
	// Create deployment
	apiObject, err := c.DatabaseV1().ArangoDeployments(ns).Create(context.Background(), depl, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)
	// Wait for deployment to be ready
	depl, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	single := depl.Status.Members.Single
	if single == nil || len(single) == 0 {
		t.Fatalf("single member is empty")
	}

	if single[0].ArangoVersion.CompareTo(expectedVersion) != 0 {
		t.Fatalf("version %s has not been set for the single member status", expectedVersion)
	}

	// Create a database client
	ctx := arangod.WithRequireAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for single server available with a valid database version
	err = waitUntilVersionUp(client, func(version driver.VersionInfo) error {
		if version.Version.CompareTo(expectedVersion) != 0 {
			t.Fatalf("database version %s is not equal expected version %s", version.Version, version)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	//Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
