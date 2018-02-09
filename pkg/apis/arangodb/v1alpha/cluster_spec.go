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

package v1alpha

import (
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArangoClusterList is a list of ArangoDB clusters.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ArangoClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArangoCluster `json:"items"`
}

// ArangoCluster contains the entire Kubernetes info for an ArangoDB cluster
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ArangoCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status"`
}

func (c *ArangoCluster) AsOwner() metav1.OwnerReference {
	controller := true
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ArangoClusterResourceKind,
		Name:       c.Name,
		UID:        c.UID,
		Controller: &controller,
	}
}

// ClusterMode specifies the type of cluster to create.
type ClusterMode string

const (
	// ClusterModeSingle yields a single server
	ClusterModeSingle ClusterMode = "single"
	// ClusterModeResilientSingle yields an agency and a resilient-single server pair
	ClusterModeResilientSingle ClusterMode = "resilientsingle"
	// ClusterModeCluster yields an full cluster (agency, dbservers & coordinators)
	ClusterModeCluster ClusterMode = "cluster"
)

// Validate the mode.
// Return errors when validation fails, nil on success.
func (m ClusterMode) Validate() error {
	switch m {
	case ClusterModeSingle, ClusterModeResilientSingle, ClusterModeCluster:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown cluster mode: '%s'", string(m)))
	}
}

// Environment in which to run the cluster
type Environment string

const (
	// EnvironmentDevelopment yields a cluster optimized for development
	EnvironmentDevelopment Environment = "development"
	// EnvironmentProduction yields a cluster optimized for production
	EnvironmentProduction Environment = "production"
)

// Validate the environment.
// Return errors when validation fails, nil on success.
func (e Environment) Validate() error {
	switch e {
	case EnvironmentDevelopment, EnvironmentProduction:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown environment: '%s'", string(e)))
	}
}

// StorageEngine specifies the type of storage engine used by the cluster
type StorageEngine string

const (
	// StorageEngineMMFiles yields a cluster using the mmfiles storage engine
	StorageEngineMMFiles StorageEngine = "mmfiles"
	// StorageEngineRocksDB yields a cluster using the rocksdb storage engine
	StorageEngineRocksDB StorageEngine = "rocksdb"
)

// Validate the storage engine.
// Return errors when validation fails, nil on success.
func (se StorageEngine) Validate() error {
	switch se {
	case StorageEngineMMFiles, StorageEngineRocksDB:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown storage engine: '%s'", string(se)))
	}
}

// validatePullPolicy the image pull policy.
// Return errors when validation fails, nil on success.
func validatePullPolicy(v v1.PullPolicy) error {
	switch v {
	case "", v1.PullAlways, v1.PullNever, v1.PullIfNotPresent:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown pull policy: '%s'", string(v)))
	}
}

// ClusterSpec contains the spec part of a Cluster resource.
type ClusterSpec struct {
	Mode            ClusterMode   `json:"mode,omitempty"`
	Environment     Environment   `json:"environment,omitempty"`
	StorageEngine   StorageEngine `json:"storageEngine,omitempty"`
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`

	RocksDB struct {
		Encryption struct {
			KeySecretName string `json:"keySecretName,omitempty"`
		} `json:"encryption"`
	} `json:"rocksdb"`

	Authentication struct {
		JWTSecretName string `json:"jwtSecretName,omitempty"`
	} `json:"auth"`

	SSL struct {
		KeySecretName    string `json:"keySecretName,omitempty"`
		OrganizationName string `json:"organizationName,omitempty"`
		ServerName       string `json:"serverName,omitempty"`
	} `json:"ssl"`
}

// SetDefaults fills in default values when a field is not specified.
func (cs *ClusterSpec) SetDefaults() {
	if cs.Mode == "" {
		cs.Mode = ClusterModeCluster
	}
	if cs.Environment == "" {
		cs.Environment = EnvironmentDevelopment
	}
	if cs.StorageEngine == "" {
		cs.StorageEngine = StorageEngineMMFiles
	}
}

// Validate the specification.
// Return errors when validation fails, nil on success.
func (cs *ClusterSpec) Validate() error {
	if err := cs.Mode.Validate(); err != nil {
		return maskAny(err)
	}
	if err := cs.Environment.Validate(); err != nil {
		return maskAny(err)
	}
	if err := cs.StorageEngine.Validate(); err != nil {
		return maskAny(err)
	}
	if err := validatePullPolicy(cs.ImagePullPolicy); err != nil {
		return maskAny(err)
	}
	return nil
}
