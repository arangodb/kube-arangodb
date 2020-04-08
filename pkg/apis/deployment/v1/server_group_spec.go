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
	"math"
	"strings"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/arangodb/kube-arangodb/pkg/util"
	arangod_options "github.com/arangodb/kube-arangodb/pkg/util/arangod/options"
	arangosync_options "github.com/arangodb/kube-arangodb/pkg/util/arangosync/options"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// ServerGroupSpec contains the specification for all servers in a specific group (e.g. all agents)
type ServerGroupSpec struct {
	// Count holds the requested number of servers
	Count *int `json:"count,omitempty"`
	// MinCount specifies a lower limit for count
	MinCount *int `json:"minCount,omitempty"`
	// MaxCount specifies a upper limit for count
	MaxCount *int `json:"maxCount,omitempty"`
	// Args holds additional commandline arguments
	Args []string `json:"args,omitempty"`
	// StorageClassName specifies the classname for storage of the servers.
	StorageClassName *string `json:"storageClassName,omitempty"`
	// Resources holds resource requests & limits
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// OverrideDetectedTotalMemory determines if memory should be overrided based on values in resources.
	OverrideDetectedTotalMemory *bool `json:"overrideDetectedTotalMemory,omitempty"`
	// Tolerations specifies the tolerations added to Pods in this group.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// Annotations specified the annotations added to Pods in this group.
	Annotations map[string]string `json:"annotations,omitempty"`
	// ServiceAccountName specifies the name of the service account used for Pods in this group.
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`
	// NodeSelector speficies a set of selectors for nodes
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Probes specifies additional behaviour for probes
	Probes *ServerGroupProbesSpec `json:"probes,omitempty"`
	// PriorityClassName specifies a priority class name
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// VolumeClaimTemplate specifies a template for volume claims
	VolumeClaimTemplate *v1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
	// VolumeResizeMode specified resize mode for pvc
	VolumeResizeMode *PVCResizeMode `json:"pvcResizeMode,omitempty"`
	// AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions
	AntiAffinity *v1.PodAntiAffinity `json:"antiAffinity,omitempty"`
	// Affinity specified additional affinity settings in ArangoDB Pod definitions
	Affinity *v1.PodAffinity `json:"affinity,omitempty"`
	// NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions
	NodeAffinity *v1.NodeAffinity `json:"nodeAffinity,omitempty"`
	// Sidecars specifies a list of additional containers to be started
	Sidecars []v1.Container `json:"sidecars,omitempty"`
	// SecurityContext specifies security context for group
	SecurityContext *ServerGroupSpecSecurityContext `json:"securityContext,omitempty"`
	// Volumes define list of volumes mounted to pod
	Volumes ServerGroupSpecVolumes `json:"volumes,omitempty"`
	// VolumeMounts define list of volume mounts mounted into server container
	VolumeMounts ServerGroupSpecVolumeMounts `json:"volumeMounts,omitempty"`
	// ExtendedRotationCheck extend checks for rotation
	ExtendedRotationCheck *bool `json:"extendedRotationCheck,omitempty"`
}

// ServerGroupSpecSecurityContext contains specification for pod security context
type ServerGroupSpecSecurityContext struct {
	// DropAllCapabilities specifies if capabilities should be dropped for this pod containers
	//
	// Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.
	DropAllCapabilities *bool `json:"dropAllCapabilities,omitempty"`
	// AddCapabilities add new capabilities to containers
	AddCapabilities []v1.Capability `json:"addCapabilities,omitempty"`
}

// GetDropAllCapabilities returns flag if capabilities should be dropped
//
// Deprecated: This function is added for backward compatibility. Will be removed in 1.1.0.
func (s *ServerGroupSpecSecurityContext) GetDropAllCapabilities() bool {
	if s == nil {
		return true
	}

	if s.DropAllCapabilities == nil {
		return true
	}

	return *s.DropAllCapabilities
}

// GetAddCapabilities add capabilities to pod context
func (s *ServerGroupSpecSecurityContext) GetAddCapabilities() []v1.Capability {
	if s == nil {
		return nil
	}

	if s.AddCapabilities == nil {
		return nil
	}

	return s.AddCapabilities
}

// NewSecurityContext creates new security context
func (s *ServerGroupSpecSecurityContext) NewSecurityContext() *v1.SecurityContext {
	r := &v1.SecurityContext{}

	capabilities := &v1.Capabilities{}

	if s.GetDropAllCapabilities() {
		capabilities.Drop = []v1.Capability{
			"ALL",
		}
	}

	if caps := s.GetAddCapabilities(); caps != nil {
		capabilities.Add = []v1.Capability{}

		capabilities.Add = append(capabilities.Add, caps...)
	}

	r.Capabilities = capabilities

	return r
}

// ServerGroupProbesSpec contains specification for probes for pods of the server group
type ServerGroupProbesSpec struct {
	// LivenessProbeDisabled if true livenessProbes are disabled
	LivenessProbeDisabled *bool `json:"livenessProbeDisabled,omitempty"`
	// LivenessProbeSpec override liveness probe configuration
	LivenessProbeSpec *ServerGroupProbeSpec `json:"livenessProbeSpec,omitempty"`

	// OldReadinessProbeDisabled if true readinessProbes are disabled
	//
	// Deprecated: This field is deprecated, keept only for backward compatibility.
	OldReadinessProbeDisabled *bool `json:"ReadinessProbeDisabled,omitempty"`
	// ReadinessProbeDisabled override flag for probe disabled in good manner (lowercase) with backward compatibility
	ReadinessProbeDisabled *bool `json:"readinessProbeDisabled,omitempty"`
	// ReadinessProbeSpec override readiness probe configuration
	ReadinessProbeSpec *ServerGroupProbeSpec `json:"readinessProbeSpec,omitempty"`
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
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
	PeriodSeconds       *int32 `json:"periodSeconds,omitempty"`
	TimeoutSeconds      *int32 `json:"timeoutSeconds,omitempty"`
	SuccessThreshold    *int32 `json:"successThreshold,omitempty"`
	FailureThreshold    *int32 `json:"failureThreshold,omitempty"`
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
func (s ServerGroupSpec) GetSidecars() []v1.Container {
	return s.Sidecars
}

// HasVolumeClaimTemplate returns whether there is a volumeClaimTemplate or not
func (s ServerGroupSpec) HasVolumeClaimTemplate() bool {
	return s.VolumeClaimTemplate != nil
}

// GetVolumeClaimTemplate returns a pointer to a volume claim template or nil if none is specified
func (s ServerGroupSpec) GetVolumeClaimTemplate() *v1.PersistentVolumeClaim {
	return s.VolumeClaimTemplate
}

// GetCount returns the value of count.
func (s ServerGroupSpec) GetCount() int {
	return util.IntOrDefault(s.Count)
}

// GetMinCount returns MinCount or 1 if not set
func (s ServerGroupSpec) GetMinCount() int {
	return util.IntOrDefault(s.MinCount, 1)
}

// GetMaxCount returns MaxCount or
func (s ServerGroupSpec) GetMaxCount() int {
	return util.IntOrDefault(s.MaxCount, math.MaxInt32)
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
		return util.StringOrDefault(pvc.Spec.StorageClassName)
	}
	return util.StringOrDefault(s.StorageClassName)
}

// GetTolerations returns the value of tolerations.
func (s ServerGroupSpec) GetTolerations() []v1.Toleration {
	return s.Tolerations
}

// GetServiceAccountName returns the value of serviceAccountName.
func (s ServerGroupSpec) GetServiceAccountName() string {
	return util.StringOrDefault(s.ServiceAccountName)
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

// Validate the given group spec
func (s ServerGroupSpec) Validate(group ServerGroup, used bool, mode DeploymentMode, env Environment) error {
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
			return maskAny(errors.Wrapf(ValidationError, "Invalid min/maxCount. Min (%d) bigger than Max (%d)", s.GetMinCount(), s.GetMaxCount()))
		}
		if s.GetCount() < s.GetMinCount() {
			return maskAny(errors.Wrapf(ValidationError, "Invalid count value %d. Expected >= %d", s.GetCount(), s.GetMinCount()))
		}
		if s.GetCount() > s.GetMaxCount() {
			return maskAny(errors.Wrapf(ValidationError, "Invalid count value %d. Expected <= %d", s.GetCount(), s.GetMaxCount()))
		}
		if s.GetCount() < minCount {
			return maskAny(errors.Wrapf(ValidationError, "Invalid count value %d. Expected >= %d (implicit minimum; by deployment mode)", s.GetCount(), minCount))
		}
		if s.GetCount() > 1 && group == ServerGroupSingle && mode == DeploymentModeSingle {
			return maskAny(errors.Wrapf(ValidationError, "Invalid count value %d. Expected 1", s.GetCount()))
		}
		if name := s.GetServiceAccountName(); name != "" {
			if err := k8sutil.ValidateOptionalResourceName(name); err != nil {
				return maskAny(errors.Wrapf(ValidationError, "Invalid serviceAccountName: %s", err))
			}
		}
		if name := s.GetStorageClassName(); name != "" {
			if err := k8sutil.ValidateOptionalResourceName(name); err != nil {
				return maskAny(errors.Wrapf(ValidationError, "Invalid storageClassName: %s", err))
			}
		}
		for _, arg := range s.Args {
			parts := strings.Split(arg, "=")
			optionKey := strings.TrimSpace(parts[0])
			if group.IsArangod() {
				if arangod_options.IsCriticalOption(optionKey) {
					return maskAny(errors.Wrapf(ValidationError, "Critical option '%s' cannot be overriden", optionKey))
				}
			} else if group.IsArangosync() {
				if arangosync_options.IsCriticalOption(optionKey) {
					return maskAny(errors.Wrapf(ValidationError, "Critical option '%s' cannot be overriden", optionKey))
				}
			}
		}

		if err := s.validate(); err != nil {
			return maskAny(err)
		}
	} else if s.GetCount() != 0 {
		return maskAny(errors.Wrapf(ValidationError, "Invalid count value %d for un-used group. Expected 0", s.GetCount()))
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
		s.validateVolumes(),
	)
}

func (s *ServerGroupSpec) validateVolumes() error {
	volumes := map[string]bool{}

	for _, volume := range s.Volumes {
		volumes[volume.Name] = true
	}

	for _, mount := range s.VolumeMounts {
		if _, ok := volumes[mount.Name]; !ok {
			return errors.Errorf("Volume %s is not defined, but required by mount", mount.Name)
		}
	}

	return nil
}

// SetDefaults fills in missing defaults
func (s *ServerGroupSpec) SetDefaults(group ServerGroup, used bool, mode DeploymentMode) {
	if s.GetCount() == 0 && used {
		switch group {
		case ServerGroupSingle:
			if mode == DeploymentModeSingle {
				s.Count = util.NewInt(1) // Single server
			} else {
				s.Count = util.NewInt(2) // ActiveFailover
			}
		default:
			s.Count = util.NewInt(3)
		}
	} else if s.GetCount() > 0 && !used {
		s.Count = nil
		s.MinCount = nil
		s.MaxCount = nil
	}
	if !s.HasVolumeClaimTemplate() {
		if _, found := s.Resources.Requests[v1.ResourceStorage]; !found {
			switch group {
			case ServerGroupSingle, ServerGroupAgents, ServerGroupDBServers:
				volumeMode := v1.PersistentVolumeFilesystem
				s.VolumeClaimTemplate = &v1.PersistentVolumeClaim{
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						VolumeMode: &volumeMode,
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("8Gi"),
							},
						},
					},
				}
			}
		}
	}
}

// setDefaultsFromResourceList fills unspecified fields with a value from given source spec.
func setDefaultsFromResourceList(s *v1.ResourceList, source v1.ResourceList) {
	for k, v := range source {
		if *s == nil {
			*s = make(v1.ResourceList)
		}
		if _, found := (*s)[k]; !found {
			(*s)[k] = v
		}
	}
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *ServerGroupSpec) SetDefaultsFrom(source ServerGroupSpec) {
	if s.Count == nil {
		s.Count = util.NewIntOrNil(source.Count)
	}
	if s.MinCount == nil {
		s.MinCount = util.NewIntOrNil(source.MinCount)
	}
	if s.MaxCount == nil {
		s.MaxCount = util.NewIntOrNil(source.MaxCount)
	}
	if s.Args == nil {
		s.Args = source.Args
	}
	if s.StorageClassName == nil {
		s.StorageClassName = util.NewStringOrNil(source.StorageClassName)
	}
	if s.Tolerations == nil {
		s.Tolerations = source.Tolerations
	}
	if s.ServiceAccountName == nil {
		s.ServiceAccountName = util.NewStringOrNil(source.ServiceAccountName)
	}
	if s.NodeSelector == nil {
		s.NodeSelector = source.NodeSelector
	}
	setDefaultsFromResourceList(&s.Resources.Limits, source.Resources.Limits)
	setDefaultsFromResourceList(&s.Resources.Requests, source.Resources.Requests)
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
			target.Count = util.NewIntOrNil(s.Count)
			resetFields = append(resetFields, fieldPrefix+".count")
		}
	}
	if s.HasVolumeClaimTemplate() != target.HasVolumeClaimTemplate() {
		target.VolumeClaimTemplate = s.GetVolumeClaimTemplate()
		resetFields = append(resetFields, fieldPrefix+".volumeClaimTemplate")
	}
	return resetFields
}
