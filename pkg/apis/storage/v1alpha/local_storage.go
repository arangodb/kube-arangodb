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

package v1alpha

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tools"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoLocalStorageList is a list of ArangoDB local storage providers.
type ArangoLocalStorageList struct {
	meta.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []ArangoLocalStorage `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoLocalStorage contains the entire Kubernetes info for an ArangoDB
// local storage provider.
type ArangoLocalStorage struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`
	Spec            LocalStorageSpec   `json:"spec"`
	Status          LocalStorageStatus `json:"status"`
}

func (d *ArangoLocalStorage) OwnerOf(in meta.Object) bool {
	return tools.IsOwner(d.AsOwner(), in)
}

// AsOwner creates an OwnerReference for the given storage
func (d *ArangoLocalStorage) AsOwner() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ArangoLocalStorageResourceKind,
		Name:       d.Name,
		UID:        d.UID,
	}
}
