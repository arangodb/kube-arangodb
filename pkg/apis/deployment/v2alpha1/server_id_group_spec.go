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

// ServerIDGroupSpec contains the specification for Image Discovery image.
type ServerIDGroupSpec struct {
	// Entrypoint overrides container executable
	Entrypoint *string `json:"entrypoint,omitempty"`
	// Tolerations specifies the tolerations added to Pods in this group.
	Tolerations []core.Toleration `json:"tolerations,omitempty"`
	// NodeSelector speficies a set of selectors for nodes
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// PriorityClassName specifies a priority class name
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// AntiAffinity specified additional antiAffinity settings in ArangoDB Pod definitions
	AntiAffinity *core.PodAntiAffinity `json:"antiAffinity,omitempty"`
	// Affinity specified additional affinity settings in ArangoDB Pod definitions
	Affinity *core.PodAffinity `json:"affinity,omitempty"`
	// NodeAffinity specified additional nodeAffinity settings in ArangoDB Pod definitions
	NodeAffinity *core.NodeAffinity `json:"nodeAffinity,omitempty"`
	// ServiceAccountName specifies the name of the service account used for Pods in this group.
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`
	// SecurityContext specifies security context for group
	SecurityContext *ServerGroupSpecSecurityContext `json:"securityContext,omitempty"`
	// Resources holds resource requests & limits
	Resources *core.ResourceRequirements `json:"resources,omitempty"`
}

func (s *ServerIDGroupSpec) Get() ServerIDGroupSpec {
	if s != nil {
		return *s
	}

	return ServerIDGroupSpec{}
}

func (s *ServerIDGroupSpec) GetServiceAccountName() string {
	if s == nil || s.ServiceAccountName == nil {
		return ""
	}

	return *s.ServiceAccountName
}

func (s *ServerIDGroupSpec) GetResources() core.ResourceRequirements {
	if s == nil || s.Resources == nil {
		return core.ResourceRequirements{
			Limits:   make(core.ResourceList),
			Requests: make(core.ResourceList),
		}
	}

	return *s.Resources
}

func (s *ServerIDGroupSpec) GetEntrypoint(defaultEntrypoint string) string {
	if s == nil || s.Entrypoint == nil {
		return defaultEntrypoint
	}

	return *s.Entrypoint
}
