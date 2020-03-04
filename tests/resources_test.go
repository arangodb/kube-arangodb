//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestResourcesChangeLimitsCluster(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	deploymentClient := kubeArangoClient.MustNewInCluster()
	ns := getNamespace(t)

	size500m, _ := resource.ParseQuantity("50m")
	size1, _ := resource.ParseQuantity("1")
	size100Gi, _ := resource.ParseQuantity("100Gi")
	size1Gi, _ := resource.ParseQuantity("1Gi")
	size2Gi, _ := resource.ParseQuantity("2Gi")

	// Prepare deployment config
	depl := newDeployment("test-chng-limits-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.DBServers.Count = util.NewInt(2)
	depl.Spec.Coordinators.Count = util.NewInt(2)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Create deployment
	_, err := deploymentClient.DatabaseV1().ArangoDeployments(ns).Create(depl)
	defer removeDeployment(deploymentClient, depl.GetName(), ns)
	assert.NoError(t, err, "failed to create deplyment: %s", err)

	testGroups := []api.ServerGroup{api.ServerGroupCoordinators, api.ServerGroupAgents, api.ServerGroupDBServers}

	testCases := []v1.ResourceRequirements{
		{
			Limits: v1.ResourceList{
				v1.ResourceCPU: size1,
			},
		},
		{
			Requests: v1.ResourceList{
				v1.ResourceCPU: size500m,
			},
		},
		{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    size500m,
				v1.ResourceMemory: size1Gi,
			},
		},
		{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    size500m,
				v1.ResourceMemory: size2Gi,
			},
		},
		{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    size1,
				v1.ResourceMemory: size100Gi,
			},
		},
	}

	for _, testgroup := range testGroups {
		t.Run(testgroup.AsRole(), func(t *testing.T) {

			_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, deploymentIsReady())
			assert.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err))

			for i, testCase := range testCases {
				t.Run(fmt.Sprintf("case-%d", i+1), func(t *testing.T) {
					depl, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
						gspec := spec.GetServerGroupSpec(testgroup)
						gspec.Resources = testCase
						spec.UpdateServerGroupSpec(testgroup, gspec)
					})
					assert.NoError(t, err, fmt.Sprintf("Failed to update deployment: %s", err))

					_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, resourcesAsRequested(kubecli, ns))
					assert.NoError(t, err, fmt.Sprintf("Deployment not rotated in time: %s", err))
				})
			}
		})
	}

}
