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

	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDeploymentList is a list of ArangoDB clusters.
type ArangoDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArangoDeployment `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoDeployment contains the entire Kubernetes info for an ArangoDB database deployment.
type ArangoDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DeploymentSpec   `json:"spec"`
	Status            DeploymentStatus `json:"status"`
}

func (c *ArangoDeployment) AsOwner() metav1.OwnerReference {
	controller := true
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ArangoDeploymentResourceKind,
		Name:       c.Name,
		UID:        c.UID,
		Controller: &controller,
	}
}

// DeploymentMode specifies the type of ArangoDB deployment to create.
type DeploymentMode string

const (
	// DeploymentModeSingle yields a single server
	DeploymentModeSingle DeploymentMode = "single"
	// DeploymentModeResilientSingle yields an agency and a resilient-single server pair
	DeploymentModeResilientSingle DeploymentMode = "resilientsingle"
	// DeploymentModeCluster yields an full cluster (agency, dbservers & coordinators)
	DeploymentModeCluster DeploymentMode = "cluster"
)

// Validate the mode.
// Return errors when validation fails, nil on success.
func (m DeploymentMode) Validate() error {
	switch m {
	case DeploymentModeSingle, DeploymentModeResilientSingle, DeploymentModeCluster:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown deployment mode: '%s'", string(m)))
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

// RocksDBSpec holds rocksdb specific configuration settings
type RocksDBSpec struct {
	Encryption struct {
		KeySecretName string `json:"keySecretName,omitempty"`
	} `json:"encryption"`
}

// Validate the given spec
func (s RocksDBSpec) Validate() error {
	if err := k8sutil.ValidateOptionalResourceName(s.Encryption.KeySecretName); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *RocksDBSpec) SetDefaults() {
	// Nothing needed
}

// AuthenticationSpec holds authentication specific configuration settings
type AuthenticationSpec struct {
	JWTSecretName string `json:"jwtSecretName,omitempty"`
}

// Validate the given spec
func (s AuthenticationSpec) Validate() error {
	if err := k8sutil.ValidateOptionalResourceName(s.JWTSecretName); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *AuthenticationSpec) SetDefaults() {
	// Nothing needed
}

// SSLSpec holds SSL specific configuration settings
type SSLSpec struct {
	KeySecretName    string `json:"keySecretName,omitempty"`
	OrganizationName string `json:"organizationName,omitempty"`
	ServerName       string `json:"serverName,omitempty"`
}

// Validate the given spec
func (s SSLSpec) Validate() error {
	if err := k8sutil.ValidateOptionalResourceName(s.KeySecretName); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *SSLSpec) SetDefaults() {
	if s.OrganizationName == "" {
		s.OrganizationName = "ArangoDB"
	}
}

type ServerGroup int

const (
	ServerGroupSingle      = 1
	ServerGroupAgents      = 2
	ServerGroupDBServers   = 3
	ServerGroupCoordinator = 4
	ServerGroupSyncMasters = 5
	ServerGroupSyncWorkers = 6
)

// ServerGroupSpec contains the specification for all servers in a specific group (e.g. all agents)
type ServerGroupSpec struct {
	// Count holds the requested number of servers
	Count int `json:"count,omitempty"`
	// Args holds additional commandline arguments
	Args []string `json:"args,omitempty"`
	// StorageClassName specifies the classname for storage of the servers.
	StorageClassName string `json:"storageClassName,omitempty"`
}

// Validate the given group spec
func (s ServerGroupSpec) Validate(group ServerGroup) error {
	if s.Count < 1 {
		return maskAny(errors.Wrapf(ValidationError, "Invalid count value %d. Expected >= 1", s.Count))
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *ServerGroupSpec) SetDefaults(group ServerGroup) {
	if s.Count == 0 {
		switch group {
		case ServerGroupSingle:
			s.Count = 1
		default:
			s.Count = 3
		}
	}
}

// DeploymentSpec contains the spec part of a ArangoDeployment resource.
type DeploymentSpec struct {
	Mode            DeploymentMode `json:"mode,omitempty"`
	Environment     Environment    `json:"environment,omitempty"`
	StorageEngine   StorageEngine  `json:"storageEngine,omitempty"`
	ImagePullPolicy v1.PullPolicy  `json:"imagePullPolicy,omitempty"`

	RocksDB        RocksDBSpec        `json:"rocksdb"`
	Authentication AuthenticationSpec `json:"auth"`
	SSL            SSLSpec            `json:"ssl"`

	Single       ServerGroupSpec `json:"single"`
	Agents       ServerGroupSpec `json:"agents"`
	DBServers    ServerGroupSpec `json:"dbservers"`
	Coordinators ServerGroupSpec `json:"coordinators"`
	SyncMasters  ServerGroupSpec `json:"syncmasters"`
	SyncWorkers  ServerGroupSpec `json:"syncworkers"`
}

// SetDefaults fills in default values when a field is not specified.
func (cs *DeploymentSpec) SetDefaults() {
	if cs.Mode == "" {
		cs.Mode = DeploymentModeCluster
	}
	if cs.Environment == "" {
		cs.Environment = EnvironmentDevelopment
	}
	if cs.StorageEngine == "" {
		cs.StorageEngine = StorageEngineMMFiles
	}
	cs.RocksDB.SetDefaults()
	cs.Authentication.SetDefaults()
	cs.SSL.SetDefaults()
	cs.Single.SetDefaults(ServerGroupSingle)
	cs.Agents.SetDefaults(ServerGroupAgents)
	cs.DBServers.SetDefaults(ServerGroupDBServers)
	cs.Coordinators.SetDefaults(ServerGroupCoordinator)
	cs.SyncMasters.SetDefaults(ServerGroupSyncMasters)
	cs.SyncWorkers.SetDefaults(ServerGroupSyncWorkers)
}

// Validate the specification.
// Return errors when validation fails, nil on success.
func (cs *DeploymentSpec) Validate() error {
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
	if err := cs.RocksDB.Validate(); err != nil {
		return maskAny(err)
	}
	if err := cs.Authentication.Validate(); err != nil {
		return maskAny(err)
	}
	if err := cs.SSL.Validate(); err != nil {
		return maskAny(err)
	}
	if err := cs.Single.Validate(ServerGroupSingle); err != nil {
		return maskAny(err)
	}
	if err := cs.Agents.Validate(ServerGroupAgents); err != nil {
		return maskAny(err)
	}
	if err := cs.DBServers.Validate(ServerGroupDBServers); err != nil {
		return maskAny(err)
	}
	if err := cs.Coordinators.Validate(ServerGroupCoordinator); err != nil {
		return maskAny(err)
	}
	if err := cs.SyncMasters.Validate(ServerGroupSyncMasters); err != nil {
		return maskAny(err)
	}
	if err := cs.SyncWorkers.Validate(ServerGroupSyncWorkers); err != nil {
		return maskAny(err)
	}
	return nil
}
