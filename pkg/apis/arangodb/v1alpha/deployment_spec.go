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
)

const (
	defaultImage = "arangodb/arangodb:latest"
)

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

// SyncSpec holds dc2dc replication specific configuration settings
type SyncSpec struct {
	Enabled         bool          `json:"enabled,omitempty"`
	Image           string        `json:"image,omitempty"`
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`

	Authentication AuthenticationSpec `json:"auth"`
	Monitoring     MonitoringSpec     `json:"monitoring"`
}

// Validate the given spec
func (s SyncSpec) Validate(mode DeploymentMode) error {
	if s.Enabled && !mode.SupportsSync() {
		return maskAny(errors.Wrapf(ValidationError, "Cannot enable sync with mode: '%s'", mode))
	}
	if s.Image == "" {
		return maskAny(errors.Wrapf(ValidationError, "image must be set"))
	}
	if err := s.Authentication.Validate(s.Enabled); err != nil {
		return maskAny(err)
	}
	if err := s.Monitoring.Validate(); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *SyncSpec) SetDefaults(defaultImage string, defaulPullPolicy v1.PullPolicy, defaultJWTSecretName string) {
	if s.Image == "" {
		s.Image = defaultImage
	}
	if s.ImagePullPolicy == "" {
		s.ImagePullPolicy = defaulPullPolicy
	}
	s.Authentication.SetDefaults(defaultJWTSecretName)
	s.Monitoring.SetDefaults()
}

// DeploymentSpec contains the spec part of a ArangoDeployment resource.
type DeploymentSpec struct {
	Mode            DeploymentMode `json:"mode,omitempty"`
	Environment     Environment    `json:"environment,omitempty"`
	StorageEngine   StorageEngine  `json:"storageEngine,omitempty"`
	Image           string         `json:"image,omitempty"`
	ImagePullPolicy v1.PullPolicy  `json:"imagePullPolicy,omitempty"`

	RocksDB        RocksDBSpec        `json:"rocksdb"`
	Authentication AuthenticationSpec `json:"auth"`
	SSL            SSLSpec            `json:"ssl"`
	Sync           SyncSpec           `json:"sync"`

	Single       ServerGroupSpec `json:"single"`
	Agents       ServerGroupSpec `json:"agents"`
	DBServers    ServerGroupSpec `json:"dbservers"`
	Coordinators ServerGroupSpec `json:"coordinators"`
	SyncMasters  ServerGroupSpec `json:"syncmasters"`
	SyncWorkers  ServerGroupSpec `json:"syncworkers"`
}

// IsAuthenticated returns true when authentication is enabled
func (s DeploymentSpec) IsAuthenticated() bool {
	return s.Authentication.IsAuthenticated()
}

// IsSecure returns true when SSL is enabled
func (s DeploymentSpec) IsSecure() bool {
	return s.SSL.KeySecretName != ""
}

// SetDefaults fills in default values when a field is not specified.
func (s *DeploymentSpec) SetDefaults(deploymentName string) {
	if s.Mode == "" {
		s.Mode = DeploymentModeCluster
	}
	if s.Environment == "" {
		s.Environment = EnvironmentDevelopment
	}
	if s.StorageEngine == "" {
		s.StorageEngine = StorageEngineMMFiles
	}
	if s.Image == "" && s.IsDevelopment() {
		s.Image = defaultImage
	}
	if s.ImagePullPolicy == "" {
		s.ImagePullPolicy = v1.PullIfNotPresent
	}
	s.RocksDB.SetDefaults()
	s.Authentication.SetDefaults(deploymentName + "-jwt")
	s.SSL.SetDefaults()
	s.Sync.SetDefaults(s.Image, s.ImagePullPolicy, deploymentName+"-sync-jwt")
	s.Single.SetDefaults(ServerGroupSingle, s.Mode.HasSingleServers(), s.Mode)
	s.Agents.SetDefaults(ServerGroupAgents, s.Mode.HasAgents(), s.Mode)
	s.DBServers.SetDefaults(ServerGroupDBServers, s.Mode.HasDBServers(), s.Mode)
	s.Coordinators.SetDefaults(ServerGroupCoordinators, s.Mode.HasCoordinators(), s.Mode)
	s.SyncMasters.SetDefaults(ServerGroupSyncMasters, s.Sync.Enabled, s.Mode)
	s.SyncWorkers.SetDefaults(ServerGroupSyncWorkers, s.Sync.Enabled, s.Mode)
}

// Validate the specification.
// Return errors when validation fails, nil on success.
func (s *DeploymentSpec) Validate() error {
	if err := s.Mode.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.mode"))
	}
	if err := s.Environment.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.environment"))
	}
	if err := s.StorageEngine.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.storageEngine"))
	}
	if err := validatePullPolicy(s.ImagePullPolicy); err != nil {
		return maskAny(errors.Wrap(err, "spec.imagePullPolicy"))
	}
	if s.Image == "" {
		return maskAny(errors.Wrapf(ValidationError, "spec.image must be set"))
	}
	if err := s.RocksDB.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.rocksdb"))
	}
	if err := s.Authentication.Validate(false); err != nil {
		return maskAny(errors.Wrap(err, "spec.auth"))
	}
	if err := s.SSL.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.ssl"))
	}
	if err := s.Sync.Validate(s.Mode); err != nil {
		return maskAny(errors.Wrap(err, "spec.sync"))
	}
	if err := s.Single.Validate(ServerGroupSingle, s.Mode.HasSingleServers(), s.Mode, s.Environment); err != nil {
		return maskAny(err)
	}
	if err := s.Agents.Validate(ServerGroupAgents, s.Mode.HasAgents(), s.Mode, s.Environment); err != nil {
		return maskAny(err)
	}
	if err := s.DBServers.Validate(ServerGroupDBServers, s.Mode.HasDBServers(), s.Mode, s.Environment); err != nil {
		return maskAny(err)
	}
	if err := s.Coordinators.Validate(ServerGroupCoordinators, s.Mode.HasCoordinators(), s.Mode, s.Environment); err != nil {
		return maskAny(err)
	}
	if err := s.SyncMasters.Validate(ServerGroupSyncMasters, s.Sync.Enabled, s.Mode, s.Environment); err != nil {
		return maskAny(err)
	}
	if err := s.SyncWorkers.Validate(ServerGroupSyncWorkers, s.Sync.Enabled, s.Mode, s.Environment); err != nil {
		return maskAny(err)
	}
	return nil
}

// IsDevelopment returns true when the spec contains a Development environment.
func (s DeploymentSpec) IsDevelopment() bool {
	return s.Environment == EnvironmentDevelopment
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s DeploymentSpec) ResetImmutableFields(target *DeploymentSpec) []string {
	var resetFields []string
	if s.Mode != target.Mode {
		target.Mode = s.Mode
		resetFields = append(resetFields, "mode")
	}
	if s.StorageEngine != target.StorageEngine {
		target.StorageEngine = s.StorageEngine
		resetFields = append(resetFields, "storageEngine")
	}
	if l := s.RocksDB.ResetImmutableFields("rocksdb", &target.RocksDB); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.Authentication.ResetImmutableFields("auth", &target.Authentication); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.Single.ResetImmutableFields(ServerGroupSingle, "single", &target.Single); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.Agents.ResetImmutableFields(ServerGroupAgents, "agents", &target.Agents); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.DBServers.ResetImmutableFields(ServerGroupDBServers, "dbservers", &target.DBServers); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.Coordinators.ResetImmutableFields(ServerGroupCoordinators, "coordinators", &target.Coordinators); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.SyncMasters.ResetImmutableFields(ServerGroupSyncMasters, "syncmasters", &target.SyncMasters); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.SyncWorkers.ResetImmutableFields(ServerGroupSyncWorkers, "syncworkers", &target.SyncWorkers); l != nil {
		resetFields = append(resetFields, l...)
	}
	return resetFields
}
