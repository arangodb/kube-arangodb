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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoClusterList is a list of ArangoDB clusters.
type ArangoClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArangoCluster `json:"items"`
}

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

type ClusterMode string

const (
	ClusterModeSingle          ClusterMode = "single"
	ClusterModeResilientSingle ClusterMode = "resilientsingle"
	ClusterModeCluster         ClusterMode = "cluster"
)

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
)

type StorageEngine string

const (
	StorageEngineMMFiles StorageEngine = "mmfiles"
	StorageEngineRocksDB StorageEngine = "rocksdb"
)

// ClusterSpec contains the spec part of a Cluster resource.
type ClusterSpec struct {
	Mode            ClusterMode        `json:"mode,omitempty"`
	Environment     Environment        `json:"environment,omitempty"`
	StorageEngine   StorageEngine      `json:"storageEngine,omitempty"`
	ImagePullPolicy v1.ImagePullPolicy `json:"imagePullPolicy,omitempty"`

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
