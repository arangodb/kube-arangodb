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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	//"github.com/arangodb/kube-arangodb/pkg/util"
)

// TODO - add description
func TestPersistence(t *testing.T) {
	longOrSkip(t)

	k8sNameSpace := getNamespace(t)
	//k8sClient := mustNewKubeClient(t)

	// volumesList, err := k8sClient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	// assert.NoError(t, err, "error while listing volumes")
	// claimsList, err := k8sClient.CoreV1().PersistentVolumeClaims(k8sNameSpace).List(metav1.ListOptions{})
	// assert.NoError(t, err, "error while listing volume claims")

	// fmt.Printf("----------------------------------------")
	// fmt.Printf("%v %v", volumesList, claimsList)
	// fmt.Printf("----------------------------------------")
	// fmt.Printf("%v %v", len(volumesList.Items), len(claimsList.Items))
	// fmt.Printf("----------------------------------------")

	mode := api.DeploymentModeCluster
	engine := api.StorageEngineRocksDB

	deploymentClient := kubeArangoClient.MustNewInCluster()
	deploymentTemplate := newDeployment(strings.Replace(fmt.Sprintf("tpers-%s-%s-%s", mode[:2], engine[:2], uniuri.NewLen(4)), ".", "", -1))
	deploymentTemplate.Spec.Mode = api.NewMode(mode)
	deploymentTemplate.Spec.StorageEngine = api.NewStorageEngine(engine)
	deploymentTemplate.Spec.TLS = api.TLSSpec{}
	//deploymentTemplate.Spec.Environment = api.NewEnvironment(api.EnvironmentDevelopment)
	//deploymentTemplate.Spec.Image = util.NewString("arangodb/arangodb:3.3.4")
	//deploymentTemplate.Spec.DBServers.Count = util.NewInt(numNodes + 1)
	deploymentTemplate.Spec.SetDefaults(deploymentTemplate.GetName()) // this must be last
	assert.NoError(t, deploymentTemplate.Spec.Validate())

	// Create deployment
	_, err := deploymentClient.DatabaseV1alpha().ArangoDeployments(k8sNameSpace).Create(deploymentTemplate)
	assert.NoError(t, err, "failed to create deplyment: %s", err)

	_, err = waitUntilDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace, deploymentIsReady())
	assert.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err)) // <-- fails here at the moment

	// TODO - add tests that check the number of volumes and claims

	// Cleanup
	removeDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace)
}
