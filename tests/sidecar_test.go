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

func TestAddSideCarToCoordinators(t *testing.T) {
	runSideCarTest(t, &sideCarTest{
		version:   "3.4.7",
		name:      "test",
		shortTest: true,
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
	container := v1.Container{Image: "nginx:1.7.9"}
	var grp = "coordinators"
	spec.AddSideCar(grp, container)

	deployment, err = updateDeployment(c, depl.GetName(), ns,
	func(depl *api.DeploymentSpec) {
		depl.Coordinators.Sidecars = spec.GroupSideCars(grp)
	})
	if err != nil {
		t.Fatalf("Failed to add ", container.Image, " to group ", grp)
	} else {
		t.Log("Updated deployment")
	}

	container = v1.Container{Image: "busybox", Command: ["sh", "-c", "sleep 3600"]}
	spec.AddSideCar(grp, container)

	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(grp)
		})
	if err != nil {
		t.Fatalf("Failed to add ", container.Image, " to group ", grp)
	} else {
		t.Log("Updated deployment")
	}

	container = v1.Container{Image: "debian", Command: "sleep 1000"}
	spec.AddSideCar(grp, container)

	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Coordinators.Sidecars = spec.GroupSideCars(grp)
		})
	if err != nil {
		t.Fatalf("Failed to add ", container.Image, " to group ", grp)
	} else {
		t.Log("Updated deployment")
	}

	removeDeployment(c, depl.GetName(), ns)

}
