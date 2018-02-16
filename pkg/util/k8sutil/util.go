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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// addOwnerRefToObject adds given owner reference to given object
func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

// LabelsForDeployment returns a map of labels, given to all resources for given deployment name
func LabelsForDeployment(deploymentName, role string) map[string]string {
	l := map[string]string{
		"arango_deployment": deploymentName,
		"app":               "arangodb",
	}
	if role != "" {
		l["role"] = role
	}
	return l
}

// DeploymentListOpt creates a ListOptions matching all labels for the given deployment name.
func DeploymentListOpt(deploymentName string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(LabelsForDeployment(deploymentName, "")).String(),
	}
}
