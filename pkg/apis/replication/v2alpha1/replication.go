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
// Author Ewout Prangsma
//

package v2alpha1

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/replication"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDeploymentReplicationList is a list of ArangoDB deployment replications.
type ArangoDeploymentReplicationList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArangoDeploymentReplication `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDeploymentReplication contains the entire Kubernetes info for an ArangoDB
// local storage provider.
type ArangoDeploymentReplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DeploymentReplicationSpec   `json:"spec"`
	Status            DeploymentReplicationStatus `json:"status"`
}

// AsOwner creates an OwnerReference for the given replication
func (d *ArangoDeploymentReplication) AsOwner() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion:         SchemeGroupVersion.String(),
		Kind:               replication.ArangoDeploymentReplicationResourceKind,
		Name:               d.Name,
		UID:                d.UID,
		Controller:         &trueVar,
		BlockOwnerDeletion: &trueVar,
	}
}
