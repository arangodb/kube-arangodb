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
// Author Lars Maier
//

package tests

import (
	"fmt"
	"testing"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func resourcesRequireRotation(wanted, given v1.ResourceRequirements) bool {
	checkList := func(wanted, given v1.ResourceList) bool {
		for k, v := range wanted {
			if gv, ok := given[k]; !ok {
				return true
			} else if v.Cmp(gv) != 0 {
				return true
			}
		}

		return false
	}

	return checkList(wanted.Limits, given.Limits) || checkList(wanted.Requests, given.Requests)
}

func TestResourcesChangeLimitsCluster(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	deploymentClient := kubeArangoClient.MustNewInCluster()
	ns := getNamespace(t)

	size500mCPU, _ := resource.ParseQuantity("50m")
	size1CPU, _ := resource.ParseQuantity("1")

	// Prepare deployment config
	depl := newDeployment("test-chng-limits-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.DBServers.Count = util.NewInt(2)
	depl.Spec.Coordinators.Count = util.NewInt(2)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Create deployment
	_, err := deploymentClient.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	defer removeDeployment(deploymentClient, depl.GetName(), ns)
	assert.NoError(t, err, "failed to create deplyment: %s", err)

	testGroups := []api.ServerGroup{api.ServerGroupCoordinators, api.ServerGroupAgents, api.ServerGroupDBServers}

	for _, testgroup := range testGroups {
		t.Run(testgroup.AsRole(), func(t *testing.T) {

			_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, deploymentIsReady())
			assert.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err))

			depl, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
				gspec := spec.GetServerGroupSpec(testgroup)
				gspec.Resources.Limits = v1.ResourceList{
					v1.ResourceCPU: size1CPU,
				}
				spec.UpdateServerGroupSpec(testgroup, gspec)
			})
			assert.NoError(t, err, fmt.Sprintf("Failed to update deployment: %s", err))

			_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, resourcesAsRequested(kubecli, ns))
			assert.NoError(t, err, fmt.Sprintf("Deployment not rotated in time: %s", err))

			depl, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
				gspec := spec.GetServerGroupSpec(testgroup)
				gspec.Resources.Requests = v1.ResourceList{
					v1.ResourceCPU: size500mCPU,
				}
				spec.UpdateServerGroupSpec(testgroup, gspec)
			})
			assert.NoError(t, err, fmt.Sprintf("Failed to update deployment: %s", err))

			_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, resourcesAsRequested(kubecli, ns))
			assert.NoError(t, err, fmt.Sprintf("Deployment not rotated in time: %s", err))
		})
	}

}
