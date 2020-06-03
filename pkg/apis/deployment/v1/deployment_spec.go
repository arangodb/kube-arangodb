//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

var (
	DefaultImage = "arangodb/arangodb:latest"
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

// DeploymentSpec contains the spec part of a ArangoDeployment resource.
type DeploymentSpec struct {
	Mode               *DeploymentMode                   `json:"mode,omitempty"`
	Environment        *Environment                      `json:"environment,omitempty"`
	StorageEngine      *StorageEngine                    `json:"storageEngine,omitempty"`
	Image              *string                           `json:"image,omitempty"`
	ImagePullPolicy    *v1.PullPolicy                    `json:"imagePullPolicy,omitempty"`
	ImagePullSecrets   []string                          `json:"imagePullSecrets,omitempty"`
	ImageDiscoveryMode *DeploymentImageDiscoveryModeSpec `json:"imageDiscoveryMode,omitempty"`
	DowntimeAllowed    *bool                             `json:"downtimeAllowed,omitempty"`
	DisableIPv6        *bool                             `json:"disableIPv6,omitempty"`

	NetworkAttachedVolumes *bool `json:"networkAttachedVolumes,omitempty"`

	// Annotations specified the annotations added to all resources
	Annotations map[string]string `json:"annotations,omitempty"`

	RestoreFrom *string `json:"restoreFrom,omitempty"`

	RestoreEncryptionSecret *string `json:"restoreEncryptionSecret,omitempty"`

	// AllowUnsafeUpgrade determines if upgrade on missing member or with not in sync shards is allowed
	AllowUnsafeUpgrade *bool `json:"allowUnsafeUpgrade,omitempty"`

	ExternalAccess ExternalAccessSpec `json:"externalAccess"`
	RocksDB        RocksDBSpec        `json:"rocksdb"`
	Authentication AuthenticationSpec `json:"auth"`
	TLS            TLSSpec            `json:"tls"`
	Sync           SyncSpec           `json:"sync"`
	License        LicenseSpec        `json:"license"`
	Metrics        MetricsSpec        `json:"metrics"`
	Lifecycle      LifecycleSpec      `json:"lifecycle,omitempty"`

	Single       ServerGroupSpec `json:"single"`
	Agents       ServerGroupSpec `json:"agents"`
	DBServers    ServerGroupSpec `json:"dbservers"`
	Coordinators ServerGroupSpec `json:"coordinators"`
	SyncMasters  ServerGroupSpec `json:"syncmasters"`
	SyncWorkers  ServerGroupSpec `json:"syncworkers"`

	Chaos ChaosSpec `json:"chaos"`

	Bootstrap BootstrapSpec `json:"bootstrap,omitempty"`
}

// GetRestoreFrom returns the restore from string or empty string if not set
func (s *DeploymentSpec) GetRestoreFrom() string {
	return util.StringOrDefault(s.RestoreFrom)
}

// HasRestoreFrom returns true if RestoreFrom is set
func (s *DeploymentSpec) HasRestoreFrom() bool {
	return s.RestoreFrom != nil
}

// Equal compares two DeploymentSpec
func (s *DeploymentSpec) Equal(other *DeploymentSpec) bool {
	return reflect.DeepEqual(s, other)
}

// GetMode returns the value of mode.
func (s DeploymentSpec) GetMode() DeploymentMode {
	return ModeOrDefault(s.Mode)
}

// GetEnvironment returns the value of environment.
func (s DeploymentSpec) GetEnvironment() Environment {
	return EnvironmentOrDefault(s.Environment)
}

// GetAnnotations returns the annotations of this group
func (s DeploymentSpec) GetAnnotations() map[string]string {
	return s.Annotations
}

// GetStorageEngine returns the value of storageEngine.
func (s DeploymentSpec) GetStorageEngine() StorageEngine {
	return StorageEngineOrDefault(s.StorageEngine)
}

// GetImage returns the value of image.
func (s DeploymentSpec) GetImage() string {
	return util.StringOrDefault(s.Image)
}

// GetSyncImage returns, if set, Sync.Image or the default image.
func (s DeploymentSpec) GetSyncImage() string {
	if s.Sync.HasSyncImage() {
		return s.Sync.GetSyncImage()
	}
	return s.GetImage()
}

// GetImagePullPolicy returns the value of imagePullPolicy.
func (s DeploymentSpec) GetImagePullPolicy() v1.PullPolicy {
	return util.PullPolicyOrDefault(s.ImagePullPolicy)
}

// IsDowntimeAllowed returns the value of downtimeAllowed.
func (s DeploymentSpec) IsDowntimeAllowed() bool {
	return util.BoolOrDefault(s.DowntimeAllowed)
}

// IsDisableIPv6 returns the value of disableIPv6.
func (s DeploymentSpec) IsDisableIPv6() bool {
	return util.BoolOrDefault(s.DisableIPv6)
}

// IsNetworkAttachedVolumes returns the value of networkAttachedVolumes, default false
func (s DeploymentSpec) IsNetworkAttachedVolumes() bool {
	return util.BoolOrDefault(s.NetworkAttachedVolumes, false)
}

// GetListenAddr returns "[::]" or "0.0.0.0" depending on IsDisableIPv6
func (s DeploymentSpec) GetListenAddr() string {
	if s.IsDisableIPv6() {
		return "0.0.0.0"
	}
	return "[::]"
}

// IsAuthenticated returns true when authentication is enabled
func (s DeploymentSpec) IsAuthenticated() bool {
	return s.Authentication.IsAuthenticated()
}

// IsSecure returns true when SSL is enabled
func (s DeploymentSpec) IsSecure() bool {
	return s.TLS.IsSecure()
}

// GetServerGroupSpec returns the server group spec (from this
// deployment spec) for the given group.
func (s DeploymentSpec) GetServerGroupSpec(group ServerGroup) ServerGroupSpec {
	switch group {
	case ServerGroupSingle:
		return s.Single
	case ServerGroupAgents:
		return s.Agents
	case ServerGroupDBServers:
		return s.DBServers
	case ServerGroupCoordinators:
		return s.Coordinators
	case ServerGroupSyncMasters:
		return s.SyncMasters
	case ServerGroupSyncWorkers:
		return s.SyncWorkers
	default:
		return ServerGroupSpec{}
	}
}

// UpdateServerGroupSpec returns the server group spec (from this
// deployment spec) for the given group.
func (s *DeploymentSpec) UpdateServerGroupSpec(group ServerGroup, gspec ServerGroupSpec) {
	switch group {
	case ServerGroupSingle:
		s.Single = gspec
	case ServerGroupAgents:
		s.Agents = gspec
	case ServerGroupDBServers:
		s.DBServers = gspec
	case ServerGroupCoordinators:
		s.Coordinators = gspec
	case ServerGroupSyncMasters:
		s.SyncMasters = gspec
	case ServerGroupSyncWorkers:
		s.SyncWorkers = gspec
	}
}

// SetDefaults fills in default values when a field is not specified.
func (s *DeploymentSpec) SetDefaults(deploymentName string) {
	if s.GetMode() == "" {
		s.Mode = NewMode(DeploymentModeCluster)
	}
	if s.GetEnvironment() == "" {
		s.Environment = NewEnvironment(EnvironmentDevelopment)
	}
	if s.GetStorageEngine() == "" {
		s.StorageEngine = NewStorageEngine(StorageEngineRocksDB)
	}
	if s.GetImage() == "" && s.IsDevelopment() {
		s.Image = util.NewString(DefaultImage)
	}
	if s.GetImagePullPolicy() == "" {
		s.ImagePullPolicy = util.NewPullPolicy(v1.PullIfNotPresent)
	}
	s.ExternalAccess.SetDefaults()
	s.RocksDB.SetDefaults()
	s.Authentication.SetDefaults(deploymentName + "-jwt")
	s.TLS.SetDefaults(deploymentName + "-ca")
	s.Sync.SetDefaults(deploymentName+"-sync-jwt", deploymentName+"-sync-client-auth-ca", deploymentName+"-sync-ca", deploymentName+"-sync-mt")
	s.Single.SetDefaults(ServerGroupSingle, s.GetMode().HasSingleServers(), s.GetMode())
	s.Agents.SetDefaults(ServerGroupAgents, s.GetMode().HasAgents(), s.GetMode())
	s.DBServers.SetDefaults(ServerGroupDBServers, s.GetMode().HasDBServers(), s.GetMode())
	s.Coordinators.SetDefaults(ServerGroupCoordinators, s.GetMode().HasCoordinators(), s.GetMode())
	s.SyncMasters.SetDefaults(ServerGroupSyncMasters, s.Sync.IsEnabled(), s.GetMode())
	s.SyncWorkers.SetDefaults(ServerGroupSyncWorkers, s.Sync.IsEnabled(), s.GetMode())
	s.Metrics.SetDefaults(deploymentName+"-exporter-jwt-token", s.Authentication.IsAuthenticated())
	s.Chaos.SetDefaults()
	s.Bootstrap.SetDefaults(deploymentName)
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *DeploymentSpec) SetDefaultsFrom(source DeploymentSpec) {
	if s.Mode == nil {
		s.Mode = NewModeOrNil(source.Mode)
	}
	if s.Environment == nil {
		s.Environment = NewEnvironmentOrNil(source.Environment)
	}
	if s.StorageEngine == nil {
		s.StorageEngine = NewStorageEngineOrNil(source.StorageEngine)
	}
	if s.Image == nil {
		s.Image = util.NewStringOrNil(source.Image)
	}
	if s.ImagePullPolicy == nil {
		s.ImagePullPolicy = util.NewPullPolicyOrNil(source.ImagePullPolicy)
	}
	if s.DowntimeAllowed == nil {
		s.DowntimeAllowed = util.NewBoolOrNil(source.DowntimeAllowed)
	}
	if s.DisableIPv6 == nil {
		s.DisableIPv6 = util.NewBoolOrNil(source.DisableIPv6)
	}

	if s.AllowUnsafeUpgrade == nil {
		s.AllowUnsafeUpgrade = util.NewBoolOrNil(source.AllowUnsafeUpgrade)
	}

	s.License.SetDefaultsFrom(source.License)
	s.ExternalAccess.SetDefaultsFrom(source.ExternalAccess)
	s.RocksDB.SetDefaultsFrom(source.RocksDB)
	s.Authentication.SetDefaultsFrom(source.Authentication)
	s.TLS.SetDefaultsFrom(source.TLS)
	s.Sync.SetDefaultsFrom(source.Sync)
	s.Single.SetDefaultsFrom(source.Single)
	s.Agents.SetDefaultsFrom(source.Agents)
	s.DBServers.SetDefaultsFrom(source.DBServers)
	s.Coordinators.SetDefaultsFrom(source.Coordinators)
	s.SyncMasters.SetDefaultsFrom(source.SyncMasters)
	s.SyncWorkers.SetDefaultsFrom(source.SyncWorkers)
	s.Metrics.SetDefaultsFrom(source.Metrics)
	s.Lifecycle.SetDefaultsFrom(source.Lifecycle)
	s.Chaos.SetDefaultsFrom(source.Chaos)
	s.Bootstrap.SetDefaultsFrom(source.Bootstrap)
}

// Validate the specification.
// Return errors when validation fails, nil on success.
func (s *DeploymentSpec) Validate() error {
	if err := s.GetMode().Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.mode"))
	}
	if err := s.GetEnvironment().Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.environment"))
	}
	if err := s.GetStorageEngine().Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.storageEngine"))
	}
	if err := validatePullPolicy(s.GetImagePullPolicy()); err != nil {
		return maskAny(errors.Wrap(err, "spec.imagePullPolicy"))
	}
	if s.GetImage() == "" {
		return maskAny(errors.Wrapf(ValidationError, "spec.image must be set"))
	}
	if err := s.ExternalAccess.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.externalAccess"))
	}
	if err := s.RocksDB.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.rocksdb"))
	}
	if err := s.Authentication.Validate(false); err != nil {
		return maskAny(errors.Wrap(err, "spec.auth"))
	}
	if err := s.TLS.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.tls"))
	}
	if err := s.Sync.Validate(s.GetMode()); err != nil {
		return maskAny(errors.Wrap(err, "spec.sync"))
	}
	if err := s.Single.Validate(ServerGroupSingle, s.GetMode().HasSingleServers(), s.GetMode(), s.GetEnvironment()); err != nil {
		return maskAny(err)
	}
	if err := s.Agents.Validate(ServerGroupAgents, s.GetMode().HasAgents(), s.GetMode(), s.GetEnvironment()); err != nil {
		return maskAny(err)
	}
	if err := s.DBServers.Validate(ServerGroupDBServers, s.GetMode().HasDBServers(), s.GetMode(), s.GetEnvironment()); err != nil {
		return maskAny(err)
	}
	if err := s.Coordinators.Validate(ServerGroupCoordinators, s.GetMode().HasCoordinators(), s.GetMode(), s.GetEnvironment()); err != nil {
		return maskAny(err)
	}
	if err := s.SyncMasters.Validate(ServerGroupSyncMasters, s.Sync.IsEnabled(), s.GetMode(), s.GetEnvironment()); err != nil {
		return maskAny(err)
	}
	if err := s.SyncWorkers.Validate(ServerGroupSyncWorkers, s.Sync.IsEnabled(), s.GetMode(), s.GetEnvironment()); err != nil {
		return maskAny(err)
	}
	if err := s.Metrics.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.metrics"))
	}
	if err := s.Chaos.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.chaos"))
	}
	if err := s.License.Validate(); err != nil {
		return maskAny(errors.Wrap(err, "spec.licenseKey"))
	}
	if err := s.Bootstrap.Validate(); err != nil {
		return maskAny(err)
	}
	return nil
}

// IsDevelopment returns true when the spec contains a Development environment.
func (s DeploymentSpec) IsDevelopment() bool {
	return s.GetEnvironment() == EnvironmentDevelopment
}

// IsProduction returns true when the spec contains a Production environment.
func (s DeploymentSpec) IsProduction() bool {
	return s.GetEnvironment() == EnvironmentProduction
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (s DeploymentSpec) ResetImmutableFields(target *DeploymentSpec) []string {
	var resetFields []string
	if s.GetMode() != target.GetMode() {
		target.Mode = NewModeOrNil(s.Mode)
		resetFields = append(resetFields, "mode")
	}
	if s.GetStorageEngine() != target.GetStorageEngine() {
		target.StorageEngine = NewStorageEngineOrNil(s.StorageEngine)
		resetFields = append(resetFields, "storageEngine")
	}
	if s.IsDisableIPv6() != target.IsDisableIPv6() {
		target.DisableIPv6 = util.NewBoolOrNil(s.DisableIPv6)
		resetFields = append(resetFields, "disableIPv6")
	}
	if l := s.ExternalAccess.ResetImmutableFields("externalAccess", &target.ExternalAccess); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.RocksDB.ResetImmutableFields("rocksdb", &target.RocksDB); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.Authentication.ResetImmutableFields("auth", &target.Authentication); l != nil {
		resetFields = append(resetFields, l...)
	}
	if l := s.Sync.ResetImmutableFields("sync", &target.Sync); l != nil {
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
	if l := s.Metrics.ResetImmutableFields("metrics", &target.Metrics); l != nil {
		resetFields = append(resetFields, l...)
	}
	return resetFields
}

// Checksum return checksum of current ArangoDeployment Spec section
func (s DeploymentSpec) Checksum() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0x", sha256.Sum256(data)), nil
}
