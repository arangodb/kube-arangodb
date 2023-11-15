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
	"math"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	arangodOptions "github.com/arangodb/kube-arangodb/pkg/util/arangod/options"
	arangosyncOptions "github.com/arangodb/kube-arangodb/pkg/util/arangosync/options"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// ServerGroupShutdownMethod enum of possible shutdown methods
type ServerGroupShutdownMethod string

// Default return default value for ServerGroupShutdownMethod
func (s *ServerGroupShutdownMethod) Default() ServerGroupShutdownMethod {
	return ServerGroupShutdownMethodAPI
}

// Get return current or default value of ServerGroupShutdownMethod
func (s *ServerGroupShutdownMethod) Get() ServerGroupShutdownMethod {
	if s == nil {
		return s.Default()
	}

	switch t := *s; t {
	case ServerGroupShutdownMethodAPI, ServerGroupShutdownMethodDelete:
		return t
	default:
		return s.Default()
	}
}

const (
	// ServerGroupShutdownMethodAPI API Shutdown method
	ServerGroupShutdownMethodAPI ServerGroupShutdownMethod = "api"
	// ServerGroupShutdownMethodDelete Pod Delete shutdown method
	ServerGroupShutdownMethodDelete ServerGroupShutdownMethod = "delete"
)

// ServerGroupSpec contains the specification for all servers in a specific group (e.g. all agents)
type ServerGroupSpec struct {
	group ServerGroup `json:"-"`

	// Count setting specifies the number of servers to start for the given group.
	// For the Agent group, this value must be a positive, odd number.
	// The default value is `3` for all groups except `single` (there the default is `1`
	// for `spec.mode: Single` and `2` for `spec.mode: ActiveFailover`).
	// For the `syncworkers` group, it is highly recommended to use the same number
	// as for the `dbservers` group.
	Count *int `json:"count,omitempty"`
	// MinCount specifies a minimum for the count of servers. If set, a specification is invalid if `count < minCount`.
	MinCount *int `json:"minCount,omitempty"`
	// MaxCount specifies a maximum for the count of servers. If set, a specification is invalid if `count > maxCount`.
	MaxCount *int `json:"maxCount,omitempty"`
	// Args setting specifies additional command-line arguments passed to all servers of this group.
	// +doc/type: []string
	// +doc/default: []
	Args []string `json:"args,omitempty"`
	// Entrypoint overrides container executable
	Entrypoint *string `json:"entrypoint,omitempty"`
	// SchedulerName define scheduler name used for group
	SchedulerName *string `json:"schedulerName,omitempty"`
	// StorageClassName specifies the classname for storage of the servers.
	StorageClassName *string `json:"storageClassName,omitempty"`
	// Resources holds resource requests & limits
	// +doc/type: core.ResourceRequirements
	// +doc/link: Documentation of core.ResourceRequirements|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core
	Resources core.ResourceRequirements `json:"resources,omitempty"`
	// OverrideDetectedTotalMemory determines if memory should be overridden based on values in resources.
	// If is set to true and Container Memory Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` to the value from the Container Memory Limits.
	// +doc/important: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` Container Environment Variable
	// +doc/default: true
	// +doc/link: Docs of the ArangoDB Envs|https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/
	OverrideDetectedTotalMemory *bool `json:"overrideDetectedTotalMemory,omitempty"`
	// MemoryReservation determines the system reservation of memory while calculating `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` value.
	// If this field is set, `ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY` is reduced by a specified value in percent.
	// Accepted Range <0, 50>. If the value is outside the accepted range, it is adjusted to the closest value.
	// +doc/default: 0
	// +doc/link: Docs of the ArangoDB Envs|https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/
	MemoryReservation *int64 `json:"memoryReservation,omitempty"`
	// OverrideDetectedNumberOfCores determines if number of cores should be overridden based on values in resources.
	// If is set to true and Container CPU Limits are set, it sets Container Environment Variable `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` to the value from the Container CPU Limits.
	// +doc/important: Values set by this feature override user-provided `ARANGODB_OVERRIDE_DETECTED_NUMBER_OF_CORES` Container Environment Variable
	// +doc/default: true
	// +doc/link: Docs of the ArangoDB Envs|https://docs.arangodb.com/devel/components/arangodb-server/environment-variables/
	OverrideDetectedNumberOfCores *bool `json:"overrideDetectedNumberOfCores,omitempty"`
	// Tolerations specifies the tolerations added to Pods in this group.
	// By default, suitable tolerations are set for the following keys with the `NoExecute` effect:
	// - `node.kubernetes.io/not-ready`
	// - `node.kubernetes.io/unreachable`
	// - `node.alpha.kubernetes.io/unreachable` (will be removed in future version)
	// For more information on tolerations, consult the https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/
	// +doc/type: []core.Toleration
	// +doc/link: Documentation of core.Toleration|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core
	Tolerations []core.Toleration `json:"tolerations,omitempty"`
	// Annotations specified the annotations added to Pods in this group.
	// Annotations are merged with `spec.annotations`.
	Annotations map[string]string `json:"annotations,omitempty"`
	// AnnotationsIgnoreList list regexp or plain definitions which annotations should be ignored
	AnnotationsIgnoreList []string `json:"annotationsIgnoreList,omitempty"`
	// AnnotationsMode Define annotations mode which should be use while overriding annotations
	AnnotationsMode *LabelsMode `json:"annotationsMode,omitempty"`
	// Labels specified the labels added to Pods in this group.
	Labels map[string]string `json:"labels,omitempty"`
	// LabelsIgnoreList list regexp or plain definitions which labels should be ignored
	LabelsIgnoreList []string `json:"labelsIgnoreList,omitempty"`
	// LabelsMode Define labels mode which should be use while overriding labels
	LabelsMode *LabelsMode `json:"labelsMode,omitempty"`
	// Envs allow to specify additional envs in this group.
	Envs ServerGroupEnvVars `json:"envs,omitempty"`
	// ServiceAccountName setting specifies the `serviceAccountName` for the `Pods` created
	// for each server of this group. If empty, it defaults to using the
	// `default` service account.
	// Using an alternative `ServiceAccount` is typically used to separate access rights.
	// The ArangoDB deployments need some very minimal access rights. With the
	// deployment of the operator, we grant the rights to 'get' all 'pod' resources.
	// If you are using a different service account, please grant these rights
	// to that service account.
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`
	// NodeSelector setting specifies a set of labels to be used as `nodeSelector` for Pods of this node.
	// +doc/type: map[string]string
	// +doc/link: Kubernetes documentation|https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Probes specifies additional behaviour for probes
	Probes *ServerGroupProbesSpec `json:"probes,omitempty"`
	// PriorityClassName specifies a priority class name
	// Will be forwarded to the pod spec.
	// +doc/link: Kubernetes documentation|https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// VolumeClaimTemplate specifies a volumeClaimTemplate used by operator to create to volume claims for pods of this group.
	// This setting is not available for group `coordinators`, `syncmasters` & `syncworkers`.
	// The default value describes a volume with `8Gi` storage, `ReadWriteOnce` access mode and volume mode set to `PersistentVolumeFilesystem`.
	// If this field is not set and `spec.<group>.resources.requests.storage` is set, then a default volume claim
	// with size as specified by `spec.<group>.resources.requests.storage` will be created. In that case `storage`
	// and `iops` is not forwarded to the pods resource requirements.
	// +doc/type: core.PersistentVolumeClaim
	// +doc/link: Documentation of core.PersistentVolumeClaim|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#persistentvolumeclaim-v1-core
	VolumeClaimTemplate *core.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
	// VolumeResizeMode specified resize mode for PVCs and PVs
	// +doc/enum: runtime|PVC will be resized in Pod runtime (EKS, GKE)
	// +doc/enum: rotate|Pod will be shutdown and PVC will be resized (AKS)
	// +doc/default: runtime
	VolumeResizeMode *PVCResizeMode `json:"pvcResizeMode,omitempty"`
	// Deprecated: VolumeAllowShrink allows shrink the volume
	VolumeAllowShrink *bool `json:"volumeAllowShrink,omitempty"`
	// AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions
	// +doc/type: core.PodAntiAffinity
	// +doc/link: Documentation of core.Pod.AntiAffinity|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podantiaffinity-v1-core
	AntiAffinity *core.PodAntiAffinity `json:"antiAffinity,omitempty"`
	// Affinity specified additional affinity settings in ArangoDB Pod definitions
	// +doc/type: core.PodAffinity
	// +doc/link: Documentation of core.PodAffinity|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podaffinity-v1-core
	Affinity *core.PodAffinity `json:"affinity,omitempty"`
	// NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions
	// +doc/type: core.NodeAffinity
	// +doc/link: Documentation of code.NodeAffinity|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#nodeaffinity-v1-core
	NodeAffinity *core.NodeAffinity `json:"nodeAffinity,omitempty"`
	// SidecarCoreNames is a list of sidecar containers which must run in the pod.
	// Some names (e.g.: "server", "worker") are reserved, and they don't have any impact.
	SidecarCoreNames []string `json:"sidecarCoreNames,omitempty"`
	// Sidecars specifies a list of additional containers to be started
	// +doc/type: []core.Container
	// +doc/link: Documentation of core.Container|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core
	Sidecars []core.Container `json:"sidecars,omitempty"`
	// SecurityContext specifies additional `securityContext` settings in ArangoDB Pod definitions.
	// This is similar (but not fully compatible) to k8s SecurityContext definition.
	// +doc/link: Kubernetes documentation|https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	SecurityContext *ServerGroupSpecSecurityContext `json:"securityContext,omitempty"`
	// Volumes define list of volumes mounted to pod
	Volumes ServerGroupSpecVolumes `json:"volumes,omitempty"`
	// VolumeMounts define list of volume mounts mounted into server container
	// +doc/type: []ServerGroupSpecVolumeMount
	// +doc/link: Documentation of ServerGroupSpecVolumeMount|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core
	VolumeMounts ServerGroupSpecVolumeMounts `json:"volumeMounts,omitempty"`
	// EphemeralVolumes keeps information about ephemeral volumes.
	EphemeralVolumes *EphemeralVolumes `json:"ephemeralVolumes,omitempty"`
	// ExtendedRotationCheck extend checks for rotation
	ExtendedRotationCheck *bool `json:"extendedRotationCheck,omitempty"`
	// InitContainers Init containers specification
	InitContainers *ServerGroupInitContainers `json:"initContainers,omitempty"`
	// ShutdownMethod describe procedure of member shutdown taken by Operator
	ShutdownMethod *ServerGroupShutdownMethod `json:"shutdownMethod,omitempty"`
	// ShutdownDelay define how long operator should delay finalizer removal after shutdown
	ShutdownDelay *int `json:"shutdownDelay,omitempty"`
	// InternalPort define port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members
	InternalPort *int `json:"internalPort,omitempty"`
	// InternalPortProtocol define protocol of port used in internal communication, can be accessed over localhost via sidecar. Only for ArangoD members
	InternalPortProtocol *ServerGroupPortProtocol `json:"internalPortProtocol,omitempty"`
	// ExternalPortEnabled if external port should be enabled. If is set to false, ports needs to be exposed via sidecar. Only for ArangoD members
	ExternalPortEnabled *bool `json:"externalPortEnabled,omitempty"`
	// AllowMemberRecreation allows to recreate member.
	// This setting changes the member recreation logic based on group:
	// - For Sync Masters, Sync Workers, Coordinator and DB-Servers it determines if a member can be recreated in case of failure (default `true`)
	// - For Agents and Single this value is hardcoded to `false` and the value provided in spec is ignored.
	AllowMemberRecreation *bool `json:"allowMemberRecreation,omitempty"`
	// TerminationGracePeriodSeconds override default TerminationGracePeriodSeconds for pods - via silent rotation
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
	// IndexMethod define group Indexing method
	// +doc/enum: random|Pick random ID for member. Enforced on the Community Operator.
	// +doc/enum: ordered|Use sequential number as Member ID, starting from 0. Enterprise Operator required.
	IndexMethod *ServerGroupIndexMethod `json:"indexMethod,omitempty"`

	// PodModes define additional modes enabled on the Pod level
	PodModes *ServerGroupSpecPodMode `json:"podModes,omitempty"`
	// Port define Port used by member
	Port *uint16 `json:"port,omitempty"`
	// ExporterPort define Port used by exporter
	ExporterPort *uint16 `json:"exporterPort,omitempty"`

	// Numactl define Numactl options passed to the process
	Numactl *ServerGroupSpecNumactl `json:"numactl,omitempty"`
}

// ServerGroupProbesSpec contains specification for probes for pods of the server group
type ServerGroupProbesSpec struct {
	// LivenessProbeDisabled if set to true, the operator does not generate a liveness probe for new pods belonging to this group
	// +doc/default: false
	LivenessProbeDisabled *bool `json:"livenessProbeDisabled,omitempty"`
	// LivenessProbeSpec override liveness probe configuration
	LivenessProbeSpec *ServerGroupProbeSpec `json:"livenessProbeSpec,omitempty"`

	// OldReadinessProbeDisabled if true readinessProbes are disabled
	//
	// Deprecated: This field is deprecated, kept only for backward compatibility.
	OldReadinessProbeDisabled *bool `json:"ReadinessProbeDisabled,omitempty"`
	// ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility
	ReadinessProbeDisabled *bool `json:"readinessProbeDisabled,omitempty"`
	// ReadinessProbeSpec override readiness probe configuration
	ReadinessProbeSpec *ServerGroupProbeSpec `json:"readinessProbeSpec,omitempty"`

	// StartupProbeDisabled if true startupProbes are disabled
	StartupProbeDisabled *bool `json:"startupProbeDisabled,omitempty"`
	// StartupProbeSpec override startup probe configuration
	StartupProbeSpec *ServerGroupProbeSpec `json:"startupProbeSpec,omitempty"`
}

// GetReadinessProbeDisabled returns in proper manner readiness probe flag with backward compatibility.
func (s ServerGroupProbesSpec) GetReadinessProbeDisabled() *bool {
	if s.OldReadinessProbeDisabled != nil {
		return s.OldReadinessProbeDisabled
	}

	return s.ReadinessProbeDisabled
}

// ServerGroupProbeSpec
type ServerGroupProbeSpec struct {
	// InitialDelaySeconds specifies number of seconds after the container has started before liveness or readiness probes are initiated.
	// Minimum value is 0.
	// +doc/default: 2
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
	// PeriodSeconds How often (in seconds) to perform the probe.
	// Minimum value is 1.
	// +doc/default: 10
	PeriodSeconds *int32 `json:"periodSeconds,omitempty"`
	// TimeoutSeconds specifies number of seconds after which the probe times out
	// Minimum value is 1.
	// +doc/default: 2
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty"`
	// SuccessThreshold Minimum consecutive successes for the probe to be considered successful after having failed.
	// Minimum value is 1.
	// +doc/default: 1
	SuccessThreshold *int32 `json:"successThreshold,omitempty"`
	// FailureThreshold when a Pod starts and the probe fails, Kubernetes will try failureThreshold times before giving up.
	// Giving up means restarting the container.
	// Minimum value is 1.
	// +doc/default: 3
	FailureThreshold *int32 `json:"failureThreshold,omitempty"`
}

// GetInitialDelaySeconds return InitialDelaySeconds valid value. In case if InitialDelaySeconds is nil default is returned.
func (s *ServerGroupProbeSpec) GetInitialDelaySeconds(d int32) int32 {
	if s == nil || s.InitialDelaySeconds == nil {
		return d // Default Kubernetes value
	}

	return *s.InitialDelaySeconds
}

// GetPeriodSeconds return PeriodSeconds valid value. In case if PeriodSeconds is nil default is returned.
func (s *ServerGroupProbeSpec) GetPeriodSeconds(d int32) int32 {
	if s == nil || s.PeriodSeconds == nil {
		return d
	}

	if *s.PeriodSeconds <= 0 {
		return 1 // Value 0 is not allowed
	}

	return *s.PeriodSeconds
}

// GetTimeoutSeconds return TimeoutSeconds valid value. In case if TimeoutSeconds is nil default is returned.
func (s *ServerGroupProbeSpec) GetTimeoutSeconds(d int32) int32 {
	if s == nil || s.TimeoutSeconds == nil {
		return d
	}

	if *s.TimeoutSeconds <= 0 {
		return 1 // Value 0 is not allowed
	}

	return *s.TimeoutSeconds
}

// GetSuccessThreshold return SuccessThreshold valid value. In case if SuccessThreshold is nil default is returned.
func (s *ServerGroupProbeSpec) GetSuccessThreshold(d int32) int32 {
	if s == nil || s.SuccessThreshold == nil {
		return d
	}

	if *s.SuccessThreshold <= 0 {
		return 1 // Value 0 is not allowed
	}

	return *s.SuccessThreshold
}

// GetFailureThreshold return FailureThreshold valid value. In case if FailureThreshold is nil default is returned.
func (s *ServerGroupProbeSpec) GetFailureThreshold(d int32) int32 {
	if s == nil || s.FailureThreshold == nil {
		return d
	}

	if *s.FailureThreshold <= 0 {
		return 1 // Value 0 is not allowed
	}

	return *s.FailureThreshold
}

// GetSidecars returns a list of sidecars the use wish to add
func (s ServerGroupSpec) GetSidecars() []core.Container {
	return s.Sidecars
}

// HasVolumeClaimTemplate returns whether there is a volumeClaimTemplate or not
func (s ServerGroupSpec) HasVolumeClaimTemplate() bool {
	return s.VolumeClaimTemplate != nil
}

// GetVolumeClaimTemplate returns a pointer to a volume claim template or nil if none is specified
func (s ServerGroupSpec) GetVolumeClaimTemplate() *core.PersistentVolumeClaim {
	return s.VolumeClaimTemplate
}

// GetCount returns the value of count.
func (s ServerGroupSpec) GetCount() int {
	return util.TypeOrDefault[int](s.Count)
}

// GetMinCount returns MinCount or 1 if not set
func (s ServerGroupSpec) GetMinCount() int {
	return util.TypeOrDefault[int](s.MinCount, 1)
}

// GetMaxCount returns MaxCount or
func (s ServerGroupSpec) GetMaxCount() int {
	return util.TypeOrDefault[int](s.MaxCount, math.MaxInt32)
}

// GetNodeSelector returns the selectors for nodes of this group
func (s ServerGroupSpec) GetNodeSelector() map[string]string {
	return s.NodeSelector
}

// GetAnnotations returns the annotations of this group
func (s ServerGroupSpec) GetAnnotations() map[string]string {
	return s.Annotations
}

// GetArgs returns the value of args.
func (s ServerGroupSpec) GetArgs() []string {
	return s.Args
}

// GetStorageClassName returns the value of storageClassName.
func (s ServerGroupSpec) GetStorageClassName() string {
	if pvc := s.GetVolumeClaimTemplate(); pvc != nil {
		return util.TypeOrDefault[string](pvc.Spec.StorageClassName)
	}
	return util.TypeOrDefault[string](s.StorageClassName)
}

// GetTolerations returns the value of tolerations.
func (s ServerGroupSpec) GetTolerations() []core.Toleration {
	return s.Tolerations
}

// GetServiceAccountName returns the value of serviceAccountName.
func (s ServerGroupSpec) GetServiceAccountName() string {
	return util.TypeOrDefault[string](s.ServiceAccountName)
}

// HasProbesSpec returns true if Probes is non nil
func (s ServerGroupSpec) HasProbesSpec() bool {
	return s.Probes != nil
}

// GetProbesSpec returns the Probes spec or the nil value if not set
func (s ServerGroupSpec) GetProbesSpec() ServerGroupProbesSpec {
	if s.HasProbesSpec() {
		return *s.Probes
	}
	return ServerGroupProbesSpec{}
}

// GetOverrideDetectedTotalMemory returns OverrideDetectedTotalMemory with default value (false)
func (s ServerGroupSpec) GetOverrideDetectedTotalMemory() bool {
	if s.OverrideDetectedTotalMemory == nil {
		return true
	}

	return *s.OverrideDetectedTotalMemory
}

// GetOverrideDetectedNumberOfCores returns OverrideDetectedNumberOfCores with default value (false)
func (s ServerGroupSpec) GetOverrideDetectedNumberOfCores() bool {
	if s.OverrideDetectedNumberOfCores == nil {
		return true
	}

	return *s.OverrideDetectedNumberOfCores
}

// Validate the given group spec
func (s ServerGroupSpec) Validate(group ServerGroup, used bool, mode DeploymentMode, env Environment) error {
	if s.group != group {
		return errors.WithStack(errors.Wrapf(ValidationError, "Group is not set"))
	}

	if used {
		minCount := 1
		if env == EnvironmentProduction {
			// Set validation boundaries for production mode
			switch group {
			case ServerGroupSingle:
				if mode == DeploymentModeActiveFailover {
					minCount = 2
				}
			case ServerGroupAgents:
				minCount = 3
			case ServerGroupDBServers, ServerGroupCoordinators, ServerGroupSyncMasters, ServerGroupSyncWorkers:
				minCount = 2
			}
		} else {
			// Set validation boundaries for development mode
			switch group {
			case ServerGroupSingle:
				if mode == DeploymentModeActiveFailover {
					minCount = 2
				}
			case ServerGroupDBServers:
				minCount = 2
			}
		}
		if s.GetMinCount() > s.GetMaxCount() {
			return errors.WithStack(errors.Wrapf(ValidationError, "Invalid min/maxCount. Min (%d) bigger than Max (%d)", s.GetMinCount(), s.GetMaxCount()))
		}
		if s.GetCount() < s.GetMinCount() {
			return errors.WithStack(errors.Wrapf(ValidationError, "Invalid count value %d. Expected >= %d", s.GetCount(), s.GetMinCount()))
		}
		if s.GetCount() > s.GetMaxCount() {
			return errors.WithStack(errors.Wrapf(ValidationError, "Invalid count value %d. Expected <= %d", s.GetCount(), s.GetMaxCount()))
		}
		if s.GetCount() < minCount {
			return errors.WithStack(errors.Wrapf(ValidationError, "Invalid count value %d. Expected >= %d (implicit minimum; by deployment mode)", s.GetCount(), minCount))
		}
		if s.GetCount() > 1 && group == ServerGroupSingle && mode == DeploymentModeSingle {
			return errors.WithStack(errors.Wrapf(ValidationError, "Invalid count value %d. Expected 1", s.GetCount()))
		}
		if name := s.GetServiceAccountName(); name != "" {
			if err := shared.ValidateOptionalResourceName(name); err != nil {
				return errors.WithStack(errors.Wrapf(ValidationError, "Invalid serviceAccountName: %s", err))
			}
		}
		if name := s.GetStorageClassName(); name != "" {
			if err := shared.ValidateOptionalResourceName(name); err != nil {
				return errors.WithStack(errors.Wrapf(ValidationError, "Invalid storageClassName: %s", err))
			}
		}
		for _, arg := range s.Args {
			parts := strings.Split(arg, "=")
			optionKey := strings.TrimSpace(parts[0])
			if group.IsArangod() {
				if arangodOptions.IsCriticalOption(optionKey) {
					return errors.WithStack(errors.Wrapf(ValidationError, "Critical option '%s' cannot be overriden", optionKey))
				}
			} else if group.IsArangosync() {
				if arangosyncOptions.IsCriticalOption(optionKey) {
					return errors.WithStack(errors.Wrapf(ValidationError, "Critical option '%s' cannot be overriden", optionKey))
				}
			}
		}

		if err := s.validate(); err != nil {
			return errors.WithStack(err)
		}
	} else if s.GetCount() != 0 {
		return errors.WithStack(errors.Wrapf(ValidationError, "Invalid count value %d for un-used group. Expected 0", s.GetCount()))
	}
	if port := s.InternalPort; port != nil {
		if err := s.InternalPortProtocol.Validate(); err != nil {
			return errors.Wrapf(err, "Validation of InternalPortProtocol failed")
		}
		switch p := *port; p {
		case 8529:
			return errors.WithStack(errors.Wrapf(ValidationError, "Port %d already in use", p))
		}
	}
	return nil
}

func (s *ServerGroupSpec) validate() error {
	if s == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("volumes", s.Volumes.Validate()),
		shared.PrefixResourceError("volumeMounts", s.VolumeMounts.Validate()),
		shared.PrefixResourceError("initContainers", s.InitContainers.Validate()),
		shared.PrefixResourceError("IndexMethod", s.IndexMethod.Validate()),
		s.validateVolumes(),
	)
}

func (s *ServerGroupSpec) validateVolumes() error {
	volumes := map[string]bool{}

	for _, volume := range s.Volumes {
		volumes[volume.Name] = true
	}

	volumes["arangod-data"] = true

	for _, mount := range s.VolumeMounts {
		if _, ok := volumes[mount.Name]; !ok {
			return errors.Newf("Volume %s is not defined, but required by mount", mount.Name)
		}
	}

	for _, container := range s.InitContainers.GetContainers() {
		for _, mount := range container.VolumeMounts {
			if _, ok := volumes[mount.Name]; !ok {
				return errors.Newf("Volume %s is not defined, but required by mount in init container %s", mount.Name, container.Name)
			}
		}
	}

	for _, container := range s.Sidecars {
		for _, mount := range s.VolumeMounts {
			if _, ok := volumes[mount.Name]; !ok {
				return errors.Newf("Volume %s is not defined, but required by mount in sidecar %s", mount.Name, container.Name)
			}
		}
	}

	return nil
}

// WithGroup copy deployment with missing group
func (s ServerGroupSpec) WithGroup(group ServerGroup) ServerGroupSpec {
	s.group = group
	return s
}

// WithDefaults copy deployment with missing defaults
func (s ServerGroupSpec) WithDefaults(group ServerGroup, used bool, mode DeploymentMode) ServerGroupSpec {
	q := s.DeepCopy()
	q.SetDefaults(group, used, mode)
	return *q
}

// SetDefaults fills in missing defaults
func (s *ServerGroupSpec) SetDefaults(group ServerGroup, used bool, mode DeploymentMode) {
	if s == nil {
		return
	}

	s.group = group

	if s.GetCount() == 0 && used {
		switch group {
		case ServerGroupSingle:
			if mode == DeploymentModeSingle {
				s.Count = util.NewType[int](1) // Single server
			} else {
				s.Count = util.NewType[int](2) // ActiveFailover
			}
		default:
			s.Count = util.NewType[int](3)
		}
	} else if s.GetCount() > 0 && !used {
		s.Count = nil
		s.MinCount = nil
		s.MaxCount = nil
	}
	if !s.HasVolumeClaimTemplate() {
		if _, found := s.Resources.Requests[core.ResourceStorage]; !found {
			switch group {
			case ServerGroupSingle, ServerGroupAgents, ServerGroupDBServers:
				volumeMode := core.PersistentVolumeFilesystem
				s.VolumeClaimTemplate = &core.PersistentVolumeClaim{
					Spec: core.PersistentVolumeClaimSpec{
						AccessModes: []core.PersistentVolumeAccessMode{
							core.ReadWriteOnce,
						},
						VolumeMode: &volumeMode,
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceStorage: resource.MustParse("8Gi"),
							},
						},
					},
				}
			}
		}
	}
}

// setStorageDefaultsFromResourceList fills unspecified storage-type fields with a value from given source spec.
func setStorageDefaultsFromResourceList(s *core.ResourceList, source core.ResourceList) {
	for k, v := range source {
		if *s == nil {
			*s = make(core.ResourceList)
		}
		if _, found := (*s)[k]; !found {
			if k != core.ResourceCPU && k != core.ResourceMemory {
				(*s)[k] = v
			}
		}
	}
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *ServerGroupSpec) SetDefaultsFrom(source ServerGroupSpec) {
	if s.Count == nil {
		s.Count = util.NewTypeOrNil[int](source.Count)
	}
	if s.MinCount == nil {
		s.MinCount = util.NewTypeOrNil[int](source.MinCount)
	}
	if s.MaxCount == nil {
		s.MaxCount = util.NewTypeOrNil[int](source.MaxCount)
	}
	if s.Args == nil {
		s.Args = source.Args
	}
	if s.StorageClassName == nil {
		s.StorageClassName = util.NewTypeOrNil[string](source.StorageClassName)
	}
	if s.Tolerations == nil {
		s.Tolerations = source.Tolerations
	}
	if s.ServiceAccountName == nil {
		s.ServiceAccountName = util.NewTypeOrNil[string](source.ServiceAccountName)
	}
	if s.NodeSelector == nil {
		s.NodeSelector = source.NodeSelector
	}
	setStorageDefaultsFromResourceList(&s.Resources.Limits, source.Resources.Limits)
	setStorageDefaultsFromResourceList(&s.Resources.Requests, source.Resources.Requests)
	if s.VolumeClaimTemplate == nil {
		s.VolumeClaimTemplate = source.VolumeClaimTemplate.DeepCopy()
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
func (s ServerGroupSpec) ResetImmutableFields(group ServerGroup, fieldPrefix string, target *ServerGroupSpec) []string {
	var resetFields []string
	if group == ServerGroupAgents {
		if s.GetCount() != target.GetCount() {
			target.Count = util.NewTypeOrNil[int](s.Count)
			resetFields = append(resetFields, fieldPrefix+".count")
		}
	}
	if s.HasVolumeClaimTemplate() != target.HasVolumeClaimTemplate() {
		target.VolumeClaimTemplate = s.GetVolumeClaimTemplate()
		resetFields = append(resetFields, fieldPrefix+".volumeClaimTemplate")
	}
	return resetFields
}

// Deprecated: GetVolumeAllowShrink returns true when it is possible to shrink the volume.
func (s ServerGroupSpec) GetVolumeAllowShrink() bool {
	if s.VolumeAllowShrink == nil {
		return false // Default value
	}

	return *s.VolumeAllowShrink
}

func (s *ServerGroupSpec) GetEntrypoint(defaultEntrypoint string) string {
	if s == nil || s.Entrypoint == nil {
		return defaultEntrypoint
	}

	return *s.Entrypoint
}

// GetShutdownDelay returns defined or default Group ShutdownDelay in seconds
func (s ServerGroupSpec) GetShutdownDelay(group ServerGroup) int {
	if s.ShutdownDelay == nil {
		switch group {
		case ServerGroupCoordinators:
			return 3
		default:
			return 0
		}
	}
	return *s.ShutdownDelay
}

// GetTerminationGracePeriod returns termination grace period as Duration
func (s ServerGroupSpec) GetTerminationGracePeriod(group ServerGroup) time.Duration {
	if v := s.TerminationGracePeriodSeconds; v == nil {
		return group.DefaultTerminationGracePeriod()
	} else {
		return time.Second * time.Duration(*v)
	}
}

// GetExternalPortEnabled returns value of ExternalPortEnabled. If ExternalPortEnabled is nil true is returned
func (s ServerGroupSpec) GetExternalPortEnabled() bool {
	if v := s.ExternalPortEnabled; v == nil {
		return true
	} else {
		return *v
	}
}

func (s *ServerGroupSpec) Group() ServerGroup {
	if s == nil {
		return ServerGroupUnknown
	}

	return s.group
}

func (s *ServerGroupSpec) GetPort() uint16 {
	if s != nil {
		if p := s.Port; p != nil {
			return *p
		}
	}

	switch s.Group() {
	case ServerGroupSyncMasters:
		return shared.ArangoSyncMasterPort
	case ServerGroupSyncWorkers:
		return shared.ArangoSyncWorkerPort
	default:
		return shared.ArangoPort
	}
}

func (s *ServerGroupSpec) GetExporterPort() uint16 {
	if s != nil {
		if p := s.ExporterPort; p != nil {
			return *p
		}
	}

	return shared.ArangoExporterPort
}

func (s *ServerGroupSpec) GetMemoryReservation() int64 {
	if s != nil {
		if v := s.MemoryReservation; v != nil {
			if q := *v; q < 0 {
				return 0
			} else if q > 50 {
				return 50
			} else {
				return q
			}
		}
	}

	return 0
}

func (s *ServerGroupSpec) CalculateMemoryReservation(memory int64) int64 {
	if r := s.GetMemoryReservation(); r > 0 {
		return int64((float64(memory)) * (float64(100-r) / 100))
	}

	return memory
}
