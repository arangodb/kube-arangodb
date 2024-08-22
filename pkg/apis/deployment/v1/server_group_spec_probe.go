//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
