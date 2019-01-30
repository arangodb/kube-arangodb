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
package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDatabaseList is a list of ArangoDB databases.
type ArangoDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArangoDatabase `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDatabase contains the entire Kubernetes info for an ArangoDB database deployment.
type ArangoDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DatabaseSpec   `json:"spec"`
	Status            ResourceStatus `json:"status"`
}

// GetDeploymentName returns the name of the deployment this database belongs to
func (ds *ArangoDatabase) GetDeploymentName() string {
	return ds.Spec.Deployment
}

// GetStatus returns the resource status of the database
func (ds *ArangoDatabase) GetStatus() *ResourceStatus {
	return &ds.Status
}

func (ds *ArangoDatabase) GetMeta() *metav1.ObjectMeta {
	return &ds.ObjectMeta
}

func (ds *ArangoDatabase) AsOwner() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ArangoDatabaseResourceKind,
		Name:       ds.Name,
		UID:        ds.UID,
	}
}

// DatabaseSpec specifies a arangodb database
type DatabaseSpec struct {
	Name       *string `json:"name,omitempty"`
	Deployment string  `json:"deployment,omitempty"`
}

// GetName returns the name of the database or empty string
func (ds *DatabaseSpec) GetName() string {
	return util.StringOrDefault(ds.Name)
}

// Validate validates a DatabaseSpec
func (ds *DatabaseSpec) Validate() error {
	return nil
}

// SetDefaults sets the default values for a DatabaseSpec
func (ds *DatabaseSpec) SetDefaults(resourceName string) {
	if ds.Name == nil {
		ds.Name = util.NewString(resourceName)
	}
}

// SetDefaultsFrom fills in the values not specified with the values form source
func (ds *DatabaseSpec) SetDefaultsFrom(source *DatabaseSpec) {
	if ds.Name == nil {
		ds.Name = util.NewStringOrNil(source.Name)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (ds *DatabaseSpec) ResetImmutableFields(target *DatabaseSpec) []string {
	var resetFields []string
	if ds.GetName() != target.GetName() {
		target.Name = util.NewStringOrNil(ds.Name)
		resetFields = append(resetFields, "Name")
	}
	return resetFields
}
