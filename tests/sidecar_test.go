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
// Author Kaveh Vahedipour <kaveh@arangodb.com>
//
package tests

import (
	"context"
	"fmt"
	"testing"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/dchest/uniuri"
	v1 "k8s.io/api/core/v1"
)

type sideCarTest struct {
	shortTest bool
	name      string
	mode      api.DeploymentMode
	version   string
	image     string
	imageTag  string
	sideCars  map[string][]v1.Container
}

type SideCarTest interface {
	IsShortTest() bool
	Mode() api.DeploymentMode
	Name() string
	Image() string
	Version() driver.Version
	GroupSideCars(string) []v1.Container
	AddSideCar(string, v1.Container)
	ClearGroupSideCars(group string)
}

func (s *sideCarTest) IsShortTest() bool {
	return s.shortTest
}
func (s *sideCarTest) Name() string {
	return s.name
}
func (s *sideCarTest) Mode() api.DeploymentMode {
	return s.mode
}
func (s *sideCarTest) Version() driver.Version {
	return driver.Version(s.version)
}
func (s *sideCarTest) GroupSideCars(group string) []v1.Container {
	if s.sideCars == nil {
		s.sideCars = make(map[string][]v1.Container)
	}
	return s.sideCars[group]
}

func (s *sideCarTest) AddSideCar(group string, container v1.Container) {
	if s.sideCars == nil {
		s.sideCars = make(map[string][]v1.Container)
	}
	s.sideCars[group] = append(s.sideCars[group], container)
}

func (s *sideCarTest) Image() string {
	imageName := "arangodb/arangodb"
	if s.image != "" {
		imageName = s.image
	}
	imageTag := "latest"
	if s.imageTag != "" {
		imageTag = s.imageTag
	}
	return fmt.Sprintf("%s:%s", imageName, imageTag)
}
func (s *sideCarTest) ClearGroupSideCars(group string) {
	s.sideCars[group] = nil
}

// TestSideCars tests side car functionality
func TestSideCars(t *testing.T) {
	runSideCarTest(t, &sideCarTest{
		version: "3.4.7",
		name:    "sidecar-tests",
	})
}

func runSideCarTest(t *testing.T, spec SideCarTest) {

	if !spec.IsShortTest() {
		longOrSkip(t)
	}

	ns := getNamespace(t)
	kubecli := mustNewKubeClient(t)
	c := kubeArangoClient.MustNewInCluster()

	depl := newDeployment(fmt.Sprintf("tu-%s-%s", spec.Name(), uniuri.NewLen(4)))
	depl.Spec.Mode = api.NewMode(spec.Mode())
	depl.Spec.TLS = api.TLSSpec{} // should auto-generate cert
	depl.Spec.Image = util.NewString(spec.Image())
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	deployment, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	deployment, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	DBClient := mustNewArangodDatabaseClient(ctx, kubecli, deployment, t, nil)
	if err := waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, spec.Version()); err != nil {
		t.Fatalf("Deployment not healthy in time: %v", err)
	}

	// Add sidecar to coordinators
	var coordinators = api.ServerGroupCoordinators.AsRole()
	var dbservers = api.ServerGroupDBServers.AsRole()
	var agents = api.ServerGroupAgents.AsRole()

	var name = "nginx"
	var image = "nginx:1.7.9"

	spec.AddSideCar(coordinators, v1.Container{Image: image, Name: name})
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
		})
	if err != nil {
		t.Fatalf("Failed to add %s to group %s", name, coordinators)
	} else {
		t.Logf("Add %s sidecar to group %s ...", name, coordinators)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	cmd1 := []string{"sh", "-c", "sleep 3600"}
	cmd2 := []string{"sh", "-c", "sleep 1800"}
	cmd := []string{"sh"}
	args := []string{"-c", "sleep 3600"}

	// Add 2nd sidecar to coordinators
	image = "busybox"
	name = "sleeper"
	spec.AddSideCar(coordinators, v1.Container{Image: image, Name: name, Command: cmd1})
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
		})
	if err != nil {
		t.Fatalf("Failed to add %s to group %s", name, coordinators)
	} else {
		t.Logf("Adding sidecar %s to group %s ...", name, coordinators)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Update command line of second sidecar
	spec.GroupSideCars(coordinators)[1].Command = cmd2
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
		})
	if err != nil {
		t.Fatalf("Failed to update %s in group %s with new command line", name, coordinators)
	} else {
		t.Logf("Update %s in group %s with new command line ...", name, coordinators)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Change command line args of second sidecar
	spec.GroupSideCars(coordinators)[1].Command = cmd
	spec.GroupSideCars(coordinators)[1].Args = args
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
		})
	if err != nil {
		t.Fatalf("Failed to update %s in group %s with new command line arguments", name, coordinators)
	} else {
		t.Logf("Updating %s in group %s with new command line arguments ...", name, coordinators)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Change environment variables of second container
	spec.GroupSideCars(coordinators)[1].Env = []v1.EnvVar{
		{Name: "Hello", Value: "World"}, {Name: "Pi", Value: "3.14159265359"}, {Name: "Two", Value: "2"}}
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
		})
	if err != nil {
		t.Fatalf("Failed to change environment variables of %s sidecars for %s", name, coordinators)
	} else {
		t.Logf("Failed to change environment variables of %s sidecars for %s", name, coordinators)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Upgrade side car image
	name = spec.GroupSideCars(coordinators)[0].Name
	spec.GroupSideCars(coordinators)[0].Image = "nginx:1.7.10"
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
		})
	if err != nil {
		t.Fatalf("Failed to update %s in group %s with new image", name, coordinators)
	} else {
		t.Logf("Update image of sidecar %s in group %s ...", name, coordinators)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Update side car image with new pull policy
	spec.GroupSideCars(coordinators)[0].ImagePullPolicy = v1.PullPolicy("Always")
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
		})
	if err != nil {
		t.Fatalf("Failed to update %s in group %s with new image pull policy", name, coordinators)
	} else {
		t.Logf("Update %s in group %s with new image pull policy ...", name, coordinators)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Remove all sidecars again
	spec.ClearGroupSideCars(coordinators)
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
		})
	if err != nil {
		t.Fatalf("Failed to remove all sidecars from group %s", coordinators)
	} else {
		t.Logf("Remove all sidecars from group %s ...", coordinators)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Adding containers to coordinators and db servers
	image = "busybox"
	name = "sleeper"
	spec.AddSideCar(coordinators, v1.Container{Image: image, Name: name, Command: cmd1})
	spec.AddSideCar(dbservers, v1.Container{Image: image, Name: name, Command: cmd1})
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
			depl.DBServers.Sidecars = spec.GroupSideCars(dbservers)
		})
	if err != nil {
		t.Fatalf("Failed to add a container to both coordinators and db servers")
	} else {
		t.Logf("Add %s sidecar to %s and %s ...", name, coordinators, dbservers)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Clear containers from both groups
	spec.ClearGroupSideCars(coordinators)
	spec.ClearGroupSideCars(dbservers)
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
			depl.DBServers.Sidecars = spec.GroupSideCars(dbservers)
		})
	if err != nil {
		t.Fatalf("Failed to delete all containers from both coordinators and db servers")
	} else {
		t.Logf("Remove all sidecars ...")
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Adding containers to agents again
	spec.AddSideCar(agents, v1.Container{Image: image, Name: name, Command: cmd1})
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
		})
	if err != nil {
		t.Fatalf("Failed to add a %s sidecar to %s", name, agents)
	} else {
		t.Logf("Failed to add a %s sidecar to %s ...", name, agents)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Clear containers from coordinators and add to db servers
	spec.ClearGroupSideCars(agents)
	spec.AddSideCar(dbservers, v1.Container{Image: image, Name: name, Command: cmd1})
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(coordinators)
			depl.DBServers.Sidecars = spec.GroupSideCars(dbservers)
		})
	if err != nil {
		t.Fatalf("Failed to delete %s containers and add %s sidecars to %s", agents, name, dbservers)
	} else {
		t.Logf("Delete %s containers and add %s sidecars to %s", agents, name, dbservers)
	}
	err = waitUntilClusterSidecarsEqualSpec(t, spec.Mode(), *depl)
	if err != nil {
		t.Fatalf("... failed: %v", err)
	} else {
		t.Log("... done")
	}

	// Clean up
	removeDeployment(c, depl.GetName(), ns)

}
