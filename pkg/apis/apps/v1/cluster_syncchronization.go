//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Jakub Wierzbowski
//

package v1

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/apps"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoClusterSynchronizationList is a list of ArangoDB jobs.
type ArangoClusterSynchronizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ArangoClusterSynchronization `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoClusterSynchronization contains definition and status of the ArangoDB type Job.
type ArangoClusterSynchronization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ArangoClusterSynchronizationSpec   `json:"spec,omitempty"`
	Status            ArangoClusterSynchronizationStatus `json:"status,omitempty"`
}

// ArangoClusterSynchronizationStatus defines the observed state of ClusterSynchronization
type ArangoClusterSynchronizationStatus struct {
	State string `json:"state,omitempty"`
}

// AsOwner creates an OwnerReference for the given job
func (a *ArangoClusterSynchronization) AsOwner() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       apps.ArangoClusterSynchronizationResourceKind,
		Name:       a.Name,
		UID:        a.UID,
		Controller: &trueVar,
	}
}
