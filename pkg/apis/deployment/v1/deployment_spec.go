//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var (
	DefaultImage = "arangodb/arangodb:latest"
)

// validatePullPolicy the image pull policy.
// Return errors when validation fails, nil on success.
func validatePullPolicy(v core.PullPolicy) error {
	switch v {
	case "", core.PullAlways, core.PullNever, core.PullIfNotPresent:
		return nil
	default:
		return errors.WithStack(errors.Wrapf(ValidationError, "Unknown pull policy: '%s'", string(v)))
	}
}

// DeploymentCommunicationMethod define communication method used for inter-cluster communication
type DeploymentCommunicationMethod string

// Get returns communication method from pointer. If pointer is nil default is returned.
func (d *DeploymentCommunicationMethod) Get() DeploymentCommunicationMethod {
	if d == nil {
		return DefaultDeploymentCommunicationMethod
	}

	switch v := *d; v {
	case DeploymentCommunicationMethodHeadlessService, DeploymentCommunicationMethodDNS, DeploymentCommunicationMethodIP, DeploymentCommunicationMethodShortDNS, DeploymentCommunicationMethodHeadlessDNS:
		return v
	default:
		return DefaultDeploymentCommunicationMethod
	}
}

// ServiceType returns Service Type for communication method
func (d *DeploymentCommunicationMethod) ServiceType() core.ServiceType {
	switch d.Get() {
	default:
		return core.ServiceTypeClusterIP
	}
}

// ServiceClusterIP returns Service ClusterIP for communication method
func (d *DeploymentCommunicationMethod) ServiceClusterIP() string {
	switch d.Get() {
	case DeploymentCommunicationMethodHeadlessDNS:
		return core.ClusterIPNone
	default:
		return ""
	}
}

// String returns string representation of method.
func (d DeploymentCommunicationMethod) String() string {
	return string(d)
}

// New returns pointer.
func (d DeploymentCommunicationMethod) New() *DeploymentCommunicationMethod {
	return &d
}

const (
	// DefaultDeploymentCommunicationMethod define default communication method.
	DefaultDeploymentCommunicationMethod = DeploymentCommunicationMethodHeadlessService
	// DeploymentCommunicationMethodHeadlessService define old communication mechanism, based on headless service.
	DeploymentCommunicationMethodHeadlessService DeploymentCommunicationMethod = "headless"
	// DeploymentCommunicationMethodDNS define ClusterIP Service DNS based communication.
	DeploymentCommunicationMethodDNS DeploymentCommunicationMethod = "dns"
	// DeploymentCommunicationMethodShortDNS define ClusterIP Service DNS based communication. Use namespaced short DNS (used in migration)
	DeploymentCommunicationMethodShortDNS DeploymentCommunicationMethod = "short-dns"
	// DeploymentCommunicationMethodHeadlessDNS define Headless Service DNS based communication.
	DeploymentCommunicationMethodHeadlessDNS DeploymentCommunicationMethod = "headless-dns"
	// DeploymentCommunicationMethodIP define ClusterIP Service IP based communication.
	DeploymentCommunicationMethodIP DeploymentCommunicationMethod = "ip"
)

// DeploymentSpec contains the spec part of a ArangoDeployment resource.
type DeploymentSpec struct {

	// Mode specifies the type of ArangoDB deployment to create.
	// +doc/enum: Cluster|Full cluster. Defaults to 3 Agents, 3 DB-Servers & 3 Coordinators.
	// +doc/enum: ActiveFailover|Active-failover single pair. Defaults to 3 Agents and 2 single servers.
	// +doc/enum: Single|Single server only (note this does not provide high availability or reliability).
	// +doc/immutable: Change of the ArangoDeployment Mode is not possible after creation.
	Mode *DeploymentMode `json:"mode,omitempty"`

	// Environment setting specifies the type of environment in which the deployment is created.
	// Possible values are:
	// - `Development` This value optimizes the deployment for development use. It is possible to run a deployment on a small number of nodes (e.g. minikube).
	// - `Production` This value optimizes the deployment for production use. It puts required affinity constraints on all pods to avoid Agents & DB-Servers from running on the same machine.
	// +doc/default: Development
	Environment *Environment `json:"environment,omitempty"`
	// StorageEngine specifies the type of storage engine used for all servers in the cluster.
	// Possible values are:
	// - `MMFiles` To use the MMFiles storage engine. Deprecated.
	// - `RocksDB` To use the RocksDB storage engine.
	// This setting cannot be changed after the cluster has been created.
	// +doc/default: RocksDB
	StorageEngine *StorageEngine `json:"storageEngine,omitempty"`
	// Image specifies the docker image to use for all ArangoDB servers.
	// In a development environment this setting defaults to arangodb/arangodb:latest.
	// For production environments this is a required setting without a default value.
	// It is highly recommend to use explicit version (not latest) for production environments.
	Image *string `json:"image,omitempty"`
	// ImagePullPolicy specifies the pull policy for the docker image to use for all ArangoDB servers.
	// +doc/type: core.PullPolicy
	// +doc/link: Documentation of core.PullPolicy|https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy
	ImagePullPolicy *core.PullPolicy `json:"imagePullPolicy,omitempty"`
	// ImagePullSecrets specifies the list of image pull secrets for the docker image to use for all ArangoDB servers.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// ImageDiscoveryMode specifies the image discovery mode.
	// +doc/enum: kubelet|Use sha256 of the discovered image in the pods
	// +doc/enum: direct|Use image provided in the spec.image directly in the pods
	ImageDiscoveryMode *DeploymentImageDiscoveryModeSpec `json:"imageDiscoveryMode,omitempty"`

	// DowntimeAllowed setting is used to allow automatic reconciliation actions that yield some downtime of the ArangoDB deployment.
	// When this setting is set to false, no automatic action that may result in downtime is allowed.
	// If the need for such an action is detected, an event is added to the ArangoDeployment.
	// Once this setting is set to true, the automatic action is executed.
	// Operations that may result in downtime are:
	// - Rotating TLS CA certificate
	// Note: It is still possible that there is some downtime when the Kubernetes cluster is down, or in a bad state, irrespective of the value of this setting.
	// +doc/default: false
	DowntimeAllowed *bool `json:"downtimeAllowed,omitempty"`
	// DisableIPv6 setting prevents the use of IPv6 addresses by ArangoDB servers.
	// This setting cannot be changed after the deployment has been created.
	// +doc/default: false
	DisableIPv6 *bool `json:"disableIPv6,omitempty"`
	// Upgrade allows to configure upgrade-related options
	// +doc/type: DeploymentUpgradeSpec
	Upgrade *DeploymentUpgradeSpec `json:"upgrade,omitempty"`
	// Features allows to configure feature flags
	// +doc/type: DeploymentFeatures
	Features *DeploymentFeatures `json:"features,omitempty"`
	// NetworkAttachedVolumes
	// If set to `true`, a ResignLeadership operation will be triggered when a DB-Server pod is evicted (rather than a CleanOutServer operation).
	// Furthermore, the pod will simply be redeployed on a different node, rather than cleaned and retired and replaced by a new member.
	// You must only set this option to true if your persistent volumes are “movable” in the sense that they can be mounted from a different k8s node, like in the case of network attached volumes.
	// If your persistent volumes are tied to a specific pod, you must leave this option on false.
	// +doc/default: false
	NetworkAttachedVolumes *bool `json:"networkAttachedVolumes,omitempty"`
	// Annotations specifies the annotations added to all ArangoDeployment owned resources (pods, services, PVC’s, PDB’s).
	Annotations map[string]string `json:"annotations,omitempty"`
	// AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored
	AnnotationsIgnoreList []string `json:"annotationsIgnoreList,omitempty"`
	// AnnotationsMode defines annotations mode which should be use while overriding annotations.
	// Possible values are:
	// - `disabled` disable annotations/labels override. Default if there is no annotations/labels set in ArangoDeployment
	// - `append` add new annotations/labels without affecting old ones
	// - `replace` replace existing annotations/labels
	AnnotationsMode *LabelsMode `json:"annotationsMode,omitempty"`
	// Labels specifies the labels added to Pods in this group.
	Labels map[string]string `json:"labels,omitempty"`
	// LabelsIgnoreList list regexp or plain definitions which labels should be ignored
	LabelsIgnoreList []string `json:"labelsIgnoreList,omitempty"`
	// LabelsMode Define labels mode which should be use while overriding labels
	// Possible values are:
	// - `disabled` disable annotations/labels override. Default if there is no annotations/labels set in ArangoDeployment
	// - `append` add new annotations/labels without affecting old ones
	// - `replace` replace existing annotations/labels
	LabelsMode *LabelsMode `json:"labelsMode,omitempty"`
	// RestoreFrom setting specifies a `ArangoBackup` resource name the cluster should be restored from.
	// After a restore or failure to do so, the status of the deployment contains information about the restore operation in the restore key.
	// It will contain some of the following fields:
	// - `requestedFrom`: name of the ArangoBackup used to restore from.
	// - `message`: optional message explaining why the restore failed.
	// - `state`: state indicating if the restore was successful or not. Possible values: Restoring, Restored, RestoreFailed
	// If the restoreFrom key is removed from the spec, the restore key is deleted as well.
	// A new restore attempt is made if and only if either in the status restore is not set or if spec.restoreFrom and status.requestedFrom are different.
	RestoreFrom *string `json:"restoreFrom,omitempty"`
	// RestoreEncryptionSecret specifies optional name of secret which contains encryption key used for restore
	RestoreEncryptionSecret *string `json:"restoreEncryptionSecret,omitempty"`
	// AllowUnsafeUpgrade determines if upgrade on missing member or with not in sync shards is allowed
	AllowUnsafeUpgrade *bool `json:"allowUnsafeUpgrade,omitempty"`
	// ExternalAccess holds configuration for the external access provided for the deployment.
	// +doc/type: ExternalAccessSpec
	ExternalAccess ExternalAccessSpec `json:"externalAccess"`
	// RocksDB holds rocksdb-specific configuration settings
	// +doc/type: RocksDBSpec
	RocksDB RocksDBSpec `json:"rocksdb"`
	// Authentication holds authentication configuration settings
	// +doc/type: AuthenticationSpec
	Authentication AuthenticationSpec `json:"auth"`
	// TLS holds TLS configuration settings
	// +doc/type: TLSSpec
	TLS TLSSpec `json:"tls"`
	// Sync holds Deployment-to-Deployment synchronization configuration settings
	// +doc/type: SyncSpec
	Sync SyncSpec `json:"sync"`
	// License holds license settings
	// +doc/type: LicenseSpec
	License LicenseSpec `json:"license"`
	// Metrics holds metrics configuration settings
	// +doc/type: MetricsSpec
	Metrics MetricsSpec `json:"metrics"`
	// Lifecycle holds lifecycle configuration settings
	// +doc/type: LifecycleSpec
	Lifecycle LifecycleSpec `json:"lifecycle,omitempty"`
	// ServerIDGroupSpec contains the specification for Image Discovery image.
	// +doc/type: ServerIDGroupSpec
	ID *ServerIDGroupSpec `json:"id,omitempty"`
	// Database holds information about database state, like maintenance mode
	// +doc/type: DatabaseSpec
	Database *DatabaseSpec `json:"database,omitempty"`
	// Single contains specification for servers running in deployment mode `Single` or `ActiveFailover`.
	// +doc/type: ServerGroupSpec
	Single ServerGroupSpec `json:"single"`
	// Agents contains specification for Agency pods running in deployment mode `Cluster` or `ActiveFailover`.
	// +doc/type: ServerGroupSpec
	Agents ServerGroupSpec `json:"agents"`
	// DBServers contains specification for DBServer pods running in deployment mode `Cluster` or `ActiveFailover`.
	// +doc/type: ServerGroupSpec
	DBServers ServerGroupSpec `json:"dbservers"`
	// Coordinators contains specification for Coordinator pods running in deployment mode `Cluster` or `ActiveFailover`.
	// +doc/type: ServerGroupSpec
	Coordinators ServerGroupSpec `json:"coordinators"`
	// SyncMasters contains specification for Syncmaster pods running in deployment mode `Cluster`.
	// +doc/type: ServerGroupSpec
	SyncMasters ServerGroupSpec `json:"syncmasters"`
	// SyncWorkers contains specification for Syncworker pods running in deployment mode `Cluster`.
	// +doc/type: ServerGroupSpec
	SyncWorkers ServerGroupSpec `json:"syncworkers"`
	// MemberPropagationMode defines how changes to pod spec should be propogated.
	// Changes to a pod’s configuration require a restart of that pod in almost all cases.
	// Pods are restarted eagerly by default, which can cause more restarts than desired, especially when updating arangod as well as the operator.
	// The propagation of the configuration changes can be deferred to the next restart, either triggered manually by the user or by another operation like an upgrade.
	// This reduces the number of restarts for upgrading both the server and the operator from two to one.
	// - `always`: Restart the member as soon as a configuration change is discovered
	// - `on-restart`: Wait until the next restart to change the member configuration
	MemberPropagationMode *DeploymentMemberPropagationMode `json:"memberPropagationMode,omitempty"`
	// ChaosSpec can be used for chaos-monkey testing of your ArangoDeployment
	// +doc/type: ChaosSpec
	Chaos ChaosSpec `json:"chaos"`
	// Recovery specifies configuration related to cluster recovery.
	// +doc/type: ArangoDeploymentRecoverySpec
	Recovery *ArangoDeploymentRecoverySpec `json:"recovery,omitempty"`
	// Bootstrap contains information for cluster bootstrapping
	// +doc/type: BootstrapSpec
	Bootstrap BootstrapSpec `json:"bootstrap,omitempty"`
	// Timeouts object allows to configure various time-outs
	// +doc/type: Timeouts
	Timeouts *Timeouts `json:"timeouts,omitempty"`
	// ClusterDomain define domain used in the kubernetes cluster.
	// Required only of domain is not set to default (cluster.local)
	// +doc/default: cluster.local
	ClusterDomain *string `json:"ClusterDomain,omitempty"`
	// CommunicationMethod define communication method used in deployment
	CommunicationMethod *DeploymentCommunicationMethod `json:"communicationMethod,omitempty"`
	// Topology define topology adjustment details, Enterprise only
	// +doc/type: TopologySpec
	Topology *TopologySpec `json:"topology,omitempty"`
	// Rebalancer defines the rebalancer specification
	// +doc/type: ArangoDeploymentRebalancerSpec
	Rebalancer *ArangoDeploymentRebalancerSpec `json:"rebalancer,omitempty"`
	// Architecture defines the list of supported architectures.
	// +doc/type: ArangoDeploymentArchitecture
	// +doc/default: ['amd64']
	Architecture ArangoDeploymentArchitecture `json:"architecture,omitempty"`
	// Timezone if specified, will set a timezone for deployment.
	// Must be in format accepted by "tzdata", e.g. `America/New_York` or `Europe/London`
	Timezone *string `json:"timezone,omitempty"`
}

// GetAllowMemberRecreation returns member recreation policy based on group and settings
func (s *DeploymentSpec) GetAllowMemberRecreation(group ServerGroup) bool {
	if s == nil {
		return false
	}

	groupSpec := s.GetServerGroupSpec(group)

	switch group {
	case ServerGroupDBServers, ServerGroupCoordinators, ServerGroupSyncMasters, ServerGroupSyncWorkers:
		if v := groupSpec.AllowMemberRecreation; v == nil {
			return true
		} else {
			return *v
		}
	default:
		return false
	}
}

// GetRestoreFrom returns the restore from string or empty string if not set
func (s *DeploymentSpec) GetRestoreFrom() string {
	return util.TypeOrDefault[string](s.RestoreFrom)
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
	return util.TypeOrDefault[string](s.Image)
}

// GetSyncImage returns, if set, Sync.Image or the default image.
func (s DeploymentSpec) GetSyncImage() string {
	if s.Sync.HasSyncImage() {
		return s.Sync.GetSyncImage()
	}
	return s.GetImage()
}

// GetImagePullPolicy returns the value of imagePullPolicy.
func (s DeploymentSpec) GetImagePullPolicy() core.PullPolicy {
	return util.TypeOrDefault[core.PullPolicy](s.ImagePullPolicy)
}

// IsDowntimeAllowed returns the value of downtimeAllowed.
func (s DeploymentSpec) IsDowntimeAllowed() bool {
	return util.TypeOrDefault[bool](s.DowntimeAllowed)
}

// IsDisableIPv6 returns the value of disableIPv6.
func (s DeploymentSpec) IsDisableIPv6() bool {
	return util.TypeOrDefault[bool](s.DisableIPv6)
}

// IsNetworkAttachedVolumes returns the value of networkAttachedVolumes, default false
func (s DeploymentSpec) IsNetworkAttachedVolumes() bool {
	return util.TypeOrDefault[bool](s.NetworkAttachedVolumes, false)
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
		return s.Single.WithGroup(group)
	case ServerGroupAgents:
		return s.Agents.WithGroup(group)
	case ServerGroupDBServers:
		return s.DBServers.WithGroup(group)
	case ServerGroupCoordinators:
		return s.Coordinators.WithGroup(group)
	case ServerGroupSyncMasters:
		return s.SyncMasters.WithGroup(group)
	case ServerGroupSyncWorkers:
		return s.SyncWorkers.WithGroup(group)
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
		s.Image = util.NewType[string](DefaultImage)
	}
	if s.GetImagePullPolicy() == "" {
		s.ImagePullPolicy = util.NewType[core.PullPolicy](core.PullIfNotPresent)
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
		s.Image = util.NewTypeOrNil[string](source.Image)
	}
	if s.ImagePullPolicy == nil {
		s.ImagePullPolicy = util.NewTypeOrNil[core.PullPolicy](source.ImagePullPolicy)
	}
	if s.DowntimeAllowed == nil {
		s.DowntimeAllowed = util.NewTypeOrNil[bool](source.DowntimeAllowed)
	}
	if s.DisableIPv6 == nil {
		s.DisableIPv6 = util.NewTypeOrNil[bool](source.DisableIPv6)
	}

	if s.AllowUnsafeUpgrade == nil {
		s.AllowUnsafeUpgrade = util.NewTypeOrNil[bool](source.AllowUnsafeUpgrade)
	}
	if s.Database == nil {
		s.Database = source.Database.DeepCopy()
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
		return errors.WithStack(errors.Wrap(err, "spec.mode"))
	}
	if err := s.GetEnvironment().Validate(); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.environment"))
	}
	if err := s.GetStorageEngine().Validate(); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.storageEngine"))
	}
	if err := validatePullPolicy(s.GetImagePullPolicy()); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.imagePullPolicy"))
	}
	if s.GetImage() == "" {
		return errors.WithStack(errors.Wrapf(ValidationError, "spec.image must be set"))
	}
	if err := s.ExternalAccess.Validate(); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.externalAccess"))
	}
	if err := s.RocksDB.Validate(); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.rocksdb"))
	}
	if err := s.Authentication.Validate(false); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.auth"))
	}
	if err := s.TLS.Validate(); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.tls"))
	}
	if err := s.Sync.Validate(s.GetMode()); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.sync"))
	}
	if err := s.Single.Validate(ServerGroupSingle, s.GetMode().HasSingleServers(), s.GetMode(), s.GetEnvironment()); err != nil {
		return errors.WithStack(err)
	}
	if err := s.Agents.Validate(ServerGroupAgents, s.GetMode().HasAgents(), s.GetMode(), s.GetEnvironment()); err != nil {
		return errors.WithStack(err)
	}
	if err := s.DBServers.Validate(ServerGroupDBServers, s.GetMode().HasDBServers(), s.GetMode(), s.GetEnvironment()); err != nil {
		return errors.WithStack(err)
	}
	if err := s.Coordinators.Validate(ServerGroupCoordinators, s.GetMode().HasCoordinators(), s.GetMode(), s.GetEnvironment()); err != nil {
		return errors.WithStack(err)
	}
	if err := s.SyncMasters.Validate(ServerGroupSyncMasters, s.Sync.IsEnabled(), s.GetMode(), s.GetEnvironment()); err != nil {
		return errors.WithStack(err)
	}
	if err := s.SyncWorkers.Validate(ServerGroupSyncWorkers, s.Sync.IsEnabled(), s.GetMode(), s.GetEnvironment()); err != nil {
		return errors.WithStack(err)
	}
	if err := s.Metrics.Validate(); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.metrics"))
	}
	if err := s.Chaos.Validate(); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.chaos"))
	}
	if err := s.License.Validate(); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.licenseKey"))
	}
	if err := s.Bootstrap.Validate(); err != nil {
		return errors.WithStack(err)
	}
	if err := s.Architecture.Validate(); err != nil {
		return errors.WithStack(errors.Wrap(err, "spec.architecture"))
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
		target.DisableIPv6 = util.NewTypeOrNil[bool](s.DisableIPv6)
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

// GetCoreContainers returns all containers' names which must running in the pod for the given group of servers.
func (s DeploymentSpec) GetCoreContainers(group ServerGroup) utils.StringList {
	groupSpec := s.GetServerGroupSpec(group)
	if len(groupSpec.SidecarCoreNames) == 0 {
		return utils.StringList{shared.ServerContainerName}
	}

	result := make(utils.StringList, 0, len(groupSpec.SidecarCoreNames)+1)
	if !utils.StringList(groupSpec.SidecarCoreNames).Has(shared.ServerContainerName) {
		result = append(result, shared.ServerContainerName)
	}
	result = append(result, groupSpec.SidecarCoreNames...)

	return result
}

func (s DeploymentSpec) GetGroupPort(group ServerGroup) uint16 {
	spec := s.GetServerGroupSpec(group)
	return spec.GetPort()
}
