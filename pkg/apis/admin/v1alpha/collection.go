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

// ArangoCollectionList is a list of ArangoDB Collections.
type ArangoCollectionList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArangoCollection `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoCollection contains the entire Kubernetes info for an ArangoDB Collection deployment.
type ArangoCollection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CollectionSpec `json:"spec"`
	Status            ResourceStatus `json:"status"`
}

// GetDeploymentName returns the name of the deployment this Collection belongs to
func (cs *ArangoCollection) GetDeploymentName() string {
	return cs.Spec.Deployment
}

// GetDatabaseName returns the name of the database this Collection belongs to
func (cs *ArangoCollection) GetDatabaseName() string {
	return cs.Spec.Database
}

// GetStatus returns the resource status of the Collection
func (ac *ArangoCollection) GetStatus() *ResourceStatus {
	return &ac.Status
}

func (ac *ArangoCollection) GetMeta() *metav1.ObjectMeta {
	return &ac.ObjectMeta
}

func (ac *ArangoCollection) AsOwner() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ArangoCollectionResourceKind,
		Name:       ac.Name,
		UID:        ac.UID,
	}
}

// CollectionType specifies the type of collection
type CollectionType string

const (
	// CollectionTypeDocument specifies a document collection
	CollectionTypeDocument CollectionType = "Document"
	// CollectionTypeEdge specifies a edge collection
	CollectionTypeEdge CollectionType = "Edge"
)

// IndexType specifies the type of an index
type IndexType string

const (
	// IndexTypeFulltext specifies Fulltext index
	IndexTypeFulltext IndexType = "Fulltext"
	// IndexTypeGeo specifies Geo index
	IndexTypeGeo IndexType = "Geo"
	// IndexTypeHash specifies Hash index
	IndexTypeHash IndexType = "Hash"
	// IndexTypePersistent specifies persistent index
	IndexTypePersistent IndexType = "Persistent"
	// IndexTypeSkipList specifies skip list
	IndexTypeSkipList IndexType = "SkipList"
)

// Index holds information about a collection index
type Index struct {
	Type IndexType `json:"type"`
}

type KeyGeneratorType string

const (
	KeyGeneratorTypeTraditional   KeyGeneratorType = "traditional"
	KeyGeneratorTypeAutoIncrement KeyGeneratorType = "autoincrement"
	KeyGeneratorTypeUUID          KeyGeneratorType = "uuid"
	KeyGeneratorTypePadded        KeyGeneratorType = "padded"
)

type KeyOptions struct {
	AllowUserKeys *bool             `json:"allowUserKeys,omitempty"`
	Type          *KeyGeneratorType `json:"type,omitempty"`
	Increment     *int              `json:"increment,omitempty"`
	Offset        *int              `json:"offset,omitempty"`
}

// CollectionSpec specifies a arangodb Collection
type CollectionSpec struct {
	Deployment           string          `json:"deployment,omitempty"`
	Database             string          `json:"database,omitempty"`
	Name                 *string         `json:"name,omitempty"`
	Type                 *CollectionType `json:"collectionType,omitempty"`
	NumberOfShards       *int            `json:"numberOfShards,omitempty"`
	ShardKeys            []string        `json:"shardKeys,omitempty"`
	ReplicationFactor    *uint           `json:"replicationFactor,omitempty"`
	WaitForSync          *bool           `json:"waitForSync,omitempty"`
	IndexBuckets         *uint           `json:"indexBuckets,omitempty"`
	IsSystem             *bool           `json:"isSystem,omitempty"`
	IsVolatile           *bool           `json:"isVolatile,omitempty"`
	DoCompact            *bool           `json:"doCompact,omitempty"`
	JournalSize          *uint           `json:"journalSize,omitempty"`
	DistributeShardsLike *string         `json:"distributeShardsLike,omitempty"`
	KeyOptions           KeyOptions      `json:"keyOptions,omitempty"`
	Indexes              []Index         `json:"indexes,omitempty"`
}

// GetName returns the name of the Collection or empty string
func (cs *CollectionSpec) GetName() string {
	return util.StringOrDefault(cs.Name)
}

// Validate validates a CollectionSpec
func (cs *CollectionSpec) Validate() error {
	return nil
}

// SetDefaults sets the default values for a CollectionSpec
func (cs *CollectionSpec) SetDefaults(resourceName string) {
	if cs.Name == nil {
		cs.Name = util.NewString(resourceName)
	}
}

// SetDefaultsFrom fills in the values not specified with the values form source
func (cs *CollectionSpec) SetDefaultsFrom(source *CollectionSpec) {
	// This is stupid work!
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (cs *CollectionSpec) ResetImmutableFields(target *CollectionSpec) []string {
	var resetFields []string
	if cs.GetName() != target.GetName() {
		target.Name = util.NewStringOrNil(cs.Name)
		resetFields = append(resetFields, "Name")
	}
	// And this too!
	return resetFields
}
