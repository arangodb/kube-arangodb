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
// Author Jan Christoph Uhde <jan@uhdejc.com>
//
package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// TestEnvironmentProduction tests if deployment comes up in production environment.
// LONG: The test ensures that the deployment fails if there are
// less nodes available than servers required.
func TestEnvironmentProduction(t *testing.T) {
	longOrSkip(t)

	mode := api.DeploymentModeCluster
	engine := api.StorageEngineRocksDB

	ns := getNamespace(t)
	kubecli := mustNewKubeClient(t)

	nodeList, err := kubecli.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Unable to receive node list: %v", err)
	}
	numNodes := len(nodeList.Items)

	c := kubeArangoClient.MustNewInCluster()
	depl := newDeployment(strings.Replace(fmt.Sprintf("tprod-%s-%s-%s", mode[:2], engine[:2], uniuri.NewLen(4)), ".", "", -1))
	depl.Spec.Mode = api.NewMode(mode)
	depl.Spec.StorageEngine = api.NewStorageEngine(engine)
	depl.Spec.TLS = api.TLSSpec{}
	depl.Spec.Environment = api.NewEnvironment(api.EnvironmentProduction)
	depl.Spec.DBServers.Count = util.NewInt(numNodes + 1)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// This test failes to validate the spec if no image is set explicitly because this is required in production mode
	if depl.Spec.Image == nil {
		depl.Spec.Image = util.NewString("arangodb/arangodb:latest")
	}
	assert.NoError(t, depl.Spec.Validate())

	dbserverCount := depl.Spec.DBServers.GetCount()
	if dbserverCount < 3 {
		t.Skipf("Not enough DBServers to run this test: server count %d", dbserverCount)
	}

	// Create deployment
	if _, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl); err != nil {
		// REVIEW - should the test already fail here
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	_, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	assert.Error(t, err, fmt.Sprintf("Deployment is up and running when it should not! There are not enough nodes(%d) for all DBServers(%d) in production modes.", numNodes, dbserverCount))

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
