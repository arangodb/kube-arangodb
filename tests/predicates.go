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
// Author Ewout Prangsma
//

package tests

import (
	"fmt"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// deploymentIsReady creates a predicate that returns nil when the deployment is in
// the running phase and the `Ready` condition is true.
func deploymentIsReady() func(*api.ArangoDeployment) error {
	return func(obj *api.ArangoDeployment) error {
		if obj.Status.Phase != api.DeploymentPhaseRunning {
			return fmt.Errorf("Expected Running phase, got %s", obj.Status.Phase)
		}
		if obj.Status.Conditions.IsTrue(api.ConditionTypeReady) {
			return nil
		}
		return fmt.Errorf("Expected Ready condition to be set, it is not")
	}
}

func resourcesAsRequested(kubecli kubernetes.Interface, ns string) func(obj *api.ArangoDeployment) error {
	return func(obj *api.ArangoDeployment) error {
		return obj.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {

			for _, m := range *status {
				pod, err := kubecli.CoreV1().Pods(ns).Get(m.PodName, metav1.GetOptions{})
				if err != nil {
					return err
				}

				c, found := k8sutil.GetContainerByName(pod, k8sutil.ServerContainerName)
				if !found {
					return fmt.Errorf("Container not found: %s", m.PodName)
				}

				if resourcesRequireRotation(spec.Resources, c.Resources) {
					return fmt.Errorf("Container of Pod %s need rotation", m.PodName)
				}
			}

			return nil
		}, nil)
	}
}
