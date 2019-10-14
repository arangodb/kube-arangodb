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

package k8sutil

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPodOwner returns the ReplicaSet that owns the given Pod.
// If the Pod has no owner of the owner is not a ReplicaSet, nil is returned.
func GetPodOwner(kubecli kubernetes.Interface, pod *v1.Pod, ns string) (*appsv1.ReplicaSet, error) {
	for _, ref := range pod.GetOwnerReferences() {
		if ref.Kind == "ReplicaSet" {
			rSets := kubecli.AppsV1().ReplicaSets(pod.GetNamespace())
			rSet, err := rSets.Get(ref.Name, metav1.GetOptions{})
			if err != nil {
				return nil, maskAny(err)
			}
			return rSet, nil
		}
	}
	return nil, nil
}

// GetReplicaSetOwner returns the Deployment that owns the given ReplicaSet.
// If the ReplicaSet has no owner of the owner is not a Deployment, nil is returned.
func GetReplicaSetOwner(kubecli kubernetes.Interface, rSet *appsv1.ReplicaSet, ns string) (*appsv1.Deployment, error) {
	for _, ref := range rSet.GetOwnerReferences() {
		if ref.Kind == "Deployment" {
			depls := kubecli.AppsV1().Deployments(rSet.GetNamespace())
			depl, err := depls.Get(ref.Name, metav1.GetOptions{})
			if err != nil {
				return nil, maskAny(err)
			}
			return depl, nil
		}
	}
	return nil, nil
}
