//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/replication"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tools"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDeploymentReplicationList is a list of ArangoDB deployment replications.
type ArangoDeploymentReplicationList struct {
	meta.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ArangoDeploymentReplication `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDeploymentReplication contains the entire Kubernetes info for an ArangoDB
// local storage provider.
type ArangoDeploymentReplication struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            DeploymentReplicationSpec   `json:"spec"`
	Status          DeploymentReplicationStatus `json:"status"`
}

func (d *ArangoDeploymentReplication) OwnerOf(in meta.Object) bool {
	return tools.IsOwner(d.AsOwner(), in)
}

// AsOwner creates an OwnerReference for the given replication
func (d *ArangoDeploymentReplication) AsOwner() meta.OwnerReference {
	trueVar := true
	return meta.OwnerReference{
		APIVersion:         SchemeGroupVersion.String(),
		Kind:               replication.ArangoDeploymentReplicationResourceKind,
		Name:               d.Name,
		UID:                d.UID,
		Controller:         &trueVar,
		BlockOwnerDeletion: &trueVar,
	}
}
