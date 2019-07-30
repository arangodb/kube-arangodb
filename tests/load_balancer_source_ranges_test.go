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
// Author Ewout Prangsma
// Author Max Neunhoeffer
//

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/dchest/uniuri"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// tests cursor forwarding with load-balanced conn., specify a source range
func TestLoadBalancingSourceRanges(t *testing.T) {
	longOrSkip(t)

	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	namePrefix := "test-lb-src-ranges-"
	depl := newDeployment(namePrefix + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.Image = util.NewString("arangodb/arangodb:latest")
	depl.Spec.ExternalAccess.Type = api.NewExternalAccessType(api.ExternalAccessTypeLoadBalancer)
	depl.Spec.ExternalAccess.LoadBalancerSourceRanges = append(depl.Spec.ExternalAccess.LoadBalancerSourceRanges, "1.2.3.0/24", "0.0.0.0/0")

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	// Prepare cleanup
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := context.Background()
	clOpts := &DatabaseClientOptions{
		UseVST:       false,
		ShortTimeout: true,
	}
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, clOpts)

	// Wait for cluster to be available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Cluster not running returning version in time: %v", err)
	}

	// Now let's use the k8s api to check if the source ranges are present in
	// the external service spec:
	svcs := kubecli.CoreV1().Services(ns)
	eaServiceName := k8sutil.CreateDatabaseExternalAccessServiceName(depl.GetName())
	// Just in case, give the service some time to appear, it should usually
	// be there already, when the deployment is ready, however, we have had
	// unstable tests in the past
	counter := 0
	var foundExternalIP string
	for {
		if svc, err := svcs.Get(eaServiceName, metav1.GetOptions{}); err == nil {
			spec := svc.Spec
			ranges := spec.LoadBalancerSourceRanges
			if len(ranges) != 2 {
				t.Errorf("LoadBalancerSourceRanges does not have length 2: %v", ranges)
			} else {
				if ranges[0] != "1.2.3.0/24" {
					t.Errorf("Expecting first LoadBalancerSourceRange to be \"1.2.3.0/24\", but ranges are: %v", ranges)
				}
				if ranges[1] != "0.0.0.0/0" {
					t.Errorf("Expecting second LoadBalancerSourceRange to be \"0.0.0.0/0\", but ranges are: %v", ranges)
				}
			}
			foundExternalIP = spec.LoadBalancerIP
			break
		}
		t.Logf("Service %s cannot be found, waiting for some time...", eaServiceName)
		time.Sleep(time.Second)
		counter++
		if counter >= 60 {
			t.Fatalf("Could not find service %s within 60 seconds, giving up.", eaServiceName)
		}
	}

	// Now change the deployment spec to use different ranges:
	depl, err = updateDeployment(c, depl.GetName(), ns,
		func(spec *api.DeploymentSpec) {
			spec.ExternalAccess.LoadBalancerSourceRanges = []string{"4.5.0.0/16"}
		})
	if err != nil {
		t.Fatalf("Failed to update the deployment")
	}

	// And check again:
	counter = 0
	for {
		time.Sleep(time.Second)
		if svc, err := svcs.Get(eaServiceName, metav1.GetOptions{}); err == nil {
			spec := svc.Spec
			ranges := spec.LoadBalancerSourceRanges
			good := true
			if len(ranges) != 1 {
				t.Logf("LoadBalancerSourceRanges does not have length 1: %v, waiting some more...", ranges)
				good = false
			} else {
				if ranges[0] != "4.5.0.0/16" {
					t.Logf("Expecting only LoadBalancerSourceRange to be \"4.5.0.0/16\", but ranges are: %v, waiting some more...", ranges)
					good = false
				} else {
					if spec.LoadBalancerIP != foundExternalIP {
						t.Errorf("Oops, the external IP of the external access service has changed: previously: %s, now: %s", foundExternalIP, spec.LoadBalancerIP)
					}
				}
			}
			if good {
				break
			}
		}
		t.Logf("Service %s cannot be found, waiting for some more time...", eaServiceName)
		counter++
		if counter >= 60 {
			t.Fatalf("Could not find changed service %s within 60 seconds, giving up.", eaServiceName)
		}
	}
	t.Logf("Success! Service %s was changed correctly.", eaServiceName)
}
