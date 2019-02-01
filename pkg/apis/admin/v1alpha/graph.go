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

// ArangoGraphList is a list of ArangoDB Graphs.
type ArangoGraphList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArangoGraph `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoGraph contains the entire Kubernetes info for an ArangoDB Graph deployment.
type ArangoGraph struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GraphSpec      `json:"spec"`
	Status            ResourceStatus `json:"status"`
}

// GetDeploymentName returns the name of the deployment this Graph belongs to
func (gs *ArangoGraph) GetDeploymentName() string {
	return gs.Spec.Deployment
}

// GetDatabaseName returns the name of the database this Graph belongs to
func (gs *ArangoGraph) GetDatabaseName() string {
	return gs.Spec.Database
}

// GetStatus returns the resource status of the Graph
func (gs *ArangoGraph) GetStatus() *ResourceStatus {
	return &gs.Status
}

func (gs *ArangoGraph) GetMeta() *metav1.ObjectMeta {
	return &gs.ObjectMeta
}

func (gs *ArangoGraph) AsOwner() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ArangoGraphResourceKind,
		Name:       gs.Name,
		UID:        gs.UID,
	}
}

type GraphOptions struct {
	SmartGraphAttribute *string `json:"smartGraphAttribute,omitempty"`
	NumberOfShards      *int    `json:"numberOfShards,omitempty"`
	ReplicationFactor   *int    `json:"replicationFactor,omitempty"`
}

// GetSmartGraphAttribute returns the smartgraphattribute or, if not set, empty string
func (gopt GraphOptions) GetSmartGraphAttribute() string {
	return util.StringOrDefault(gopt.SmartGraphAttribute)
}

// GetNumberOfShards returns the number of shards or, if not set, zero
func (gopt GraphOptions) GetNumberOfShards() int {
	return util.IntOrDefault(gopt.NumberOfShards)
}

// GetReplicationFactor returns the replication factor, it not set, zero
func (gopt GraphOptions) GetReplicationFactor() int {
	return util.IntOrDefault(gopt.ReplicationFactor)
}

// GraphSpec specifies a arangodb Graph
type GraphSpec struct {
	Name       *string `json:"name,omitempty"`
	Deployment string  `json:"deployment,omitempty"`
	Database   string  `json:"database,omitempty"`
	IsSmart    *bool   `json:"isSmart,omitempty"`
	//EdgeDefinitions   []driver.EdgeDefinition `json:"edgeDefinitions,omitempty"`
	Options           GraphOptions `json:"options,omitempty"`
	OrphanCollections []string     `json:"orphanCollections,omitempty"`
}

// GetName returns the name of the Graph or empty string
func (gs *GraphSpec) GetName() string {
	return util.StringOrDefault(gs.Name)
}

func (gs *GraphSpec) GetIsSmart() bool {
	return util.BoolOrDefault(gs.IsSmart)
}

func (gs *GraphSpec) GetDeploymentName() string {
	return gs.Deployment
}

func (gs *GraphSpec) GetOptions() GraphOptions {
	return gs.Options
}

// Validate validates a GraphSpec
func (gs *GraphSpec) Validate() error {
	return nil
}

// SetDefaults sets the default values for a GraphSpec
func (gs *GraphSpec) SetDefaults(resourceName string) {
	if gs.Name == nil {
		gs.Name = util.NewString(resourceName)
	}
}

// SetDefaultsFrom fills in the values not specified with the values form source
func (gs *GraphSpec) SetDefaultsFrom(source *GraphSpec) {
	if gs.Name == nil {
		gs.Name = util.NewStringOrNil(source.Name)
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (gs *GraphSpec) ResetImmutableFields(target *GraphSpec) []string {
	var resetFields []string
	if gs.GetName() != target.GetName() {
		target.Name = util.NewStringOrNil(gs.Name)
		resetFields = append(resetFields, "Name")
	}
	return resetFields
}
