//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/permission"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoPermissionRoleUserBindingList is a list of ArangoPermissionRoleUserBinding.
type ArangoPermissionRoleUserBindingList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []ArangoPermissionRoleUserBinding `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoPermissionRoleUserBinding binds a Role to a User with a scope Policy within an ArangoDeployment.
type ArangoPermissionRoleUserBinding struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArangoPermissionRoleUserBindingSpec   `json:"spec"`
	Status ArangoPermissionRoleUserBindingStatus `json:"status"`
}

// AsOwner creates an OwnerReference for the given resource
func (a *ArangoPermissionRoleUserBinding) AsOwner() meta.OwnerReference {
	trueVar := true
	return meta.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       permission.ArangoPermissionRoleUserBindingResourceKind,
		Name:       a.Name,
		UID:        a.UID,
		Controller: &trueVar,
	}
}

func (a *ArangoPermissionRoleUserBinding) GetStatus() ArangoPermissionRoleUserBindingStatus {
	return a.Status
}

func (a *ArangoPermissionRoleUserBinding) SetStatus(status ArangoPermissionRoleUserBindingStatus) {
	a.Status = status
}

func (a *ArangoPermissionRoleUserBinding) Ready() bool {
	if a == nil {
		return false
	}

	if !a.Status.Conditions.IsTrue(ReadyCondition) {
		return false
	}

	return true
}
