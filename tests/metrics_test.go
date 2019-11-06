//
// DISCLAIMER
//
// Copyright 2019 ArangoDB GmbH, Cologne, Germany
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
// Author Max Neunhoeffer
//
package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/dchest/uniuri"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

func TestAddingMetrics(t *testing.T) {
	longOrSkip(t)

	ns := getNamespace(t)
	kubecli := mustNewKubeClient(t)
	c := kubeArangoClient.MustNewInCluster()

	depl := newDeployment(fmt.Sprintf("%s-%s", "arangodb-metrics-test", uniuri.NewLen(4)))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineRocksDB)
	depl.Spec.TLS = api.TLSSpec{}         // should auto-generate cert
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

	// Try to switch on metrics:
	expectedResourceRequirement := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("100m"),
		},
	}
	deployment, err = updateDeployment(c, depl.GetName(), ns, func(depl *api.DeploymentSpec) {
		depl.Metrics = api.MetricsSpec{
			Enabled:   util.NewBool(true),
			Image:     util.NewString("arangodb/arangodb-exporter:0.1.6"),
			Resources: expectedResourceRequirement,
		}
	})
	require.NoErrorf(t, err, "Failed to add metrics")
	t.Log("Updated deployment by adding metrics")

	var resourcesRequirementsExporterCheck api.ServerGroupFunc = func(group api.ServerGroup, spec api.ServerGroupSpec,
		status *api.MemberStatusList) error {

		if !group.IsExportMetrics() {
			return nil
		}

		for _, m := range *status {

			pod, err := kubecli.CoreV1().Pods(ns).Get(m.PodName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			exporter, found := k8sutil.GetContainerByName(pod, k8sutil.ExporterContainerName)
			if !found {
				return fmt.Errorf("expected exporter to be enabled")
			}

			if k8sutil.IsResourceRequirementsChanged(expectedResourceRequirement, exporter.Resources) {
				return fmt.Errorf("resources have not been changed: expected %v, actual %v",
					expectedResourceRequirement, exporter.Resources)
			}
		}
		return nil
	}
	_, err = waitUntilDeploymentMembers(c, deployment.GetName(), ns, resourcesRequirementsExporterCheck, 7*time.Minute)
	require.NoError(t, err)

	expectedResourceRequirement.Requests[v1.ResourceCPU] = resource.MustParse("110m")
	deployment, err = updateDeployment(c, depl.GetName(), ns, func(depl *api.DeploymentSpec) {
		depl.Metrics.Resources = expectedResourceRequirement
	})
	require.NoErrorf(t, err, "failed to change resource requirements for metrics")
	t.Log("Updated deployment by changing metrics")
	_, err = waitUntilDeploymentMembers(c, deployment.GetName(), ns, resourcesRequirementsExporterCheck, 7*time.Minute)
	require.NoError(t, err)

	err = waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, "")
	require.NoErrorf(t, err, "Deployment not healthy in time")
	t.Log("Deployment healthy")

	_, err = waitUntilService(kubecli, depl.GetName()+"-exporter", ns,
		func(service *corev1.Service) error {
			return nil
		}, time.Second*30)
	require.NoErrorf(t, err, "Exporter service did not show up in time")
	t.Log("Found exporter service")

	_, err = waitUntilEndpoints(kubecli, depl.GetName()+"-exporter", ns,
		func(endpoints *corev1.Endpoints) error {
			count := 0
			for _, subset := range endpoints.Subsets {
				count += len(subset.Addresses)
			}
			t.Logf("Found %d endpoints in the Endpoints resource", count)
			if count < 6 {
				return errors.New("did not find enough endpoints in Endpoints resource")
			}
			return nil
		}, time.Second*360) // needs a full rotation with extra containers
	require.NoErrorf(t, err, "Exporter endpoints did not show up in time")
	t.Log("Found exporter endpoints")

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
