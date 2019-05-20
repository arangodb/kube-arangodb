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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/dchest/uniuri"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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
	if err := waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, ""); err != nil {
		t.Fatalf("Deployment not healthy in time: %v", err)
	}

	// Try to switch on metrics:
	deployment, err = updateDeployment(c, depl.GetName(), ns,
		func(depl *api.DeploymentSpec) {
			depl.Metrics = api.MetricsSpec{
				Enabled: util.NewBool(true),
				Image:   util.NewString("arangodb/arangodb-exporter:0.1.6"),
			}
		})
	if err != nil {
		t.Fatalf("Failed to add metrics")
	} else {
		t.Log("Updated deployment by adding metrics")
	}

	if err := waitUntilArangoDeploymentHealthy(deployment, DBClient, kubecli, ""); err != nil {
		t.Errorf("Deployment not healthy in time: %v", err)
	} else {
		t.Log("Deployment healthy")
	}

	_, err = waitUntilService(kubecli, depl.GetName()+"-exporter", ns,
		func(service *corev1.Service) error {
			return nil
		}, time.Second*30)
	if err != nil {
		t.Errorf("Exporter service did not show up in time")
	} else {
		t.Log("Found exporter service")
	}

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
	if err != nil {
		t.Errorf("Exporter endpoints did not show up in time")
	} else {
		t.Log("Found exporter endpoints")
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
