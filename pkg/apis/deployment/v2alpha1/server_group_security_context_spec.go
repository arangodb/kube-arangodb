//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import core "k8s.io/api/core/v1"

// ServerGroupSpecSecurityContext contains specification for pod security context
type ServerGroupSpecSecurityContext struct {
	// DropAllCapabilities specifies if capabilities should be dropped for this pod containers
	//
	// Deprecated: This field is added for backward compatibility. Will be removed in 1.1.0.
	DropAllCapabilities *bool `json:"dropAllCapabilities,omitempty"`
	// AddCapabilities add new capabilities to containers
	AddCapabilities []core.Capability `json:"addCapabilities,omitempty"`

	AllowPrivilegeEscalation *bool  `json:"allowPrivilegeEscalation,omitempty"`
	Privileged               *bool  `json:"privileged,omitempty"`
	ReadOnlyRootFilesystem   *bool  `json:"readOnlyRootFilesystem,omitempty"`
	RunAsNonRoot             *bool  `json:"runAsNonRoot,omitempty"`
	RunAsUser                *int64 `json:"runAsUser,omitempty"`
	RunAsGroup               *int64 `json:"runAsGroup,omitempty"`

	SupplementalGroups []int64 `json:"supplementalGroups,omitempty"`
	FSGroup            *int64  `json:"fsGroup,omitempty"`

	SeccompProfile *core.SeccompProfile `json:"seccompProfile,omitempty" protobuf:"bytes,11,opt,name=seccompProfile"`
	SELinuxOptions *core.SELinuxOptions `json:"seLinuxOptions,omitempty" protobuf:"bytes,3,opt,name=seLinuxOptions"`
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
func (s *ServerGroupSpecSecurityContext) GetAddCapabilities() []core.Capability {
	if s == nil {
		return nil
	}

	return s.AddCapabilities
}

// NewSecurityContext creates new pod security context
func (s *ServerGroupSpecSecurityContext) NewPodSecurityContext() *core.PodSecurityContext {
	if s == nil {
		return nil
	}

	if s.FSGroup == nil && len(s.SupplementalGroups) == 0 {
		return nil
	}

	return &core.PodSecurityContext{
		SupplementalGroups: s.SupplementalGroups,
		FSGroup:            s.FSGroup,
	}
}

// NewSecurityContext creates new security context
func (s *ServerGroupSpecSecurityContext) NewSecurityContext() *core.SecurityContext {
	r := &core.SecurityContext{}

	if s != nil {
		r.AllowPrivilegeEscalation = s.AllowPrivilegeEscalation
		r.Privileged = s.Privileged
		r.ReadOnlyRootFilesystem = s.ReadOnlyRootFilesystem
		r.RunAsNonRoot = s.RunAsNonRoot
		r.RunAsUser = s.RunAsUser
		r.RunAsGroup = s.RunAsGroup

		r.SeccompProfile = s.SeccompProfile.DeepCopy()
		r.SELinuxOptions = s.SELinuxOptions.DeepCopy()
	}

	capabilities := &core.Capabilities{}

	if s.GetDropAllCapabilities() {
		capabilities.Drop = []core.Capability{
			"ALL",
		}
	}

	if caps := s.GetAddCapabilities(); caps != nil {
		capabilities.Add = []core.Capability{}

		capabilities.Add = append(capabilities.Add, caps...)
	}

	r.Capabilities = capabilities

	return r
}
