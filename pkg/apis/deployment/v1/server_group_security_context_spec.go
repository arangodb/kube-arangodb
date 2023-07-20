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
	"sort"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

const (
	defaultRunAsUser  = 1000
	defaultRunAsGroup = 2000
	defaultFSGroup    = 3000
)

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

	// Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported
	// sysctls (by the container runtime) might fail to launch.
	// Map Value can be String or Int
	// +doc/example: sysctls:
	// +doc/example:   "kernel.shm_rmid_forced": "0"
	// +doc/example:   "net.core.somaxconn": 1024
	// +doc/example:   "kernel.msgmax": "65536"
	// +doc/type: map[string]intstr.IntOrString
	// +doc/link: Documentation|https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/
	Sysctls map[string]intstr.IntOrString `json:"sysctls,omitempty"`

	// SeccompProfile defines a pod/container's seccomp profile settings. Only one profile source may be set.
	// +doc/type: core.SeccompProfile
	// +doc/link: Documentation of core.SeccompProfile|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#seccompprofile-v1-core
	SeccompProfile *core.SeccompProfile `json:"seccompProfile,omitempty" protobuf:"bytes,11,opt,name=seccompProfile"`

	// SELinuxOptions are the labels to be applied to the container
	// +doc/type: core.SELinuxOptions
	// +doc/link: Documentation of core.SELinuxOptions|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#selinuxoptions-v1-core
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

// NewPodSecurityContext creates new pod security context
func (s *ServerGroupSpecSecurityContext) NewPodSecurityContext(secured bool) *core.PodSecurityContext {
	var psc *core.PodSecurityContext
	if s != nil && (s.FSGroup != nil || len(s.SupplementalGroups) > 0) {
		psc = &core.PodSecurityContext{
			SupplementalGroups: s.SupplementalGroups,
			FSGroup:            s.FSGroup,
		}
	}

	if s != nil && len(s.Sysctls) > 0 {
		var sysctls []core.Sysctl
		for k, v := range s.Sysctls {
			sysctls = append(sysctls, core.Sysctl{
				Name:  k,
				Value: v.String(),
			})
		}

		sort.Slice(sysctls, func(i, j int) bool {
			return sysctls[i].Name < sysctls[j].Name
		})

		if psc == nil {
			psc = &core.PodSecurityContext{}
		}

		psc.Sysctls = sysctls
	}

	if secured {
		if psc == nil {
			psc = &core.PodSecurityContext{}
		}

		if psc.FSGroup == nil {
			psc.FSGroup = util.NewType[int64](defaultFSGroup)
		}
	}

	return psc
}

// NewSecurityContext creates new security context
func (s *ServerGroupSpecSecurityContext) NewSecurityContext(secured ...bool) *core.SecurityContext {
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

	if len(secured) > 0 && secured[0] {
		if r.RunAsUser == nil {
			r.RunAsUser = util.NewType[int64](defaultRunAsUser)
		}
		if r.RunAsGroup == nil {
			r.RunAsGroup = util.NewType[int64](defaultRunAsGroup)
		}
		if r.RunAsNonRoot == nil {
			r.RunAsNonRoot = util.NewType[bool](true)
		}
		if r.ReadOnlyRootFilesystem == nil {
			r.ReadOnlyRootFilesystem = util.NewType[bool](true)
		}

		if capabilities.Drop == nil {
			capabilities.Drop = []core.Capability{
				"ALL",
			}
		}
	}

	r.Capabilities = capabilities

	return r
}
