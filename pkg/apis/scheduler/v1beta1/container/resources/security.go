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

package resources

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/interfaces"
)

var _ interfaces.Container[Security] = &Security{}

type Security struct {
	// SecurityContext holds container-level security attributes and common container settings.
	// +doc/type: core.SecurityContext
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	SecurityContext *core.SecurityContext `json:"securityContext,omitempty"`
}

func (s *Security) Apply(_ *core.PodTemplateSpec, template *core.Container) error {
	if s == nil {
		return nil
	}

	template.SecurityContext = s.SecurityContext.DeepCopy()

	return nil
}

func (s *Security) With(other *Security) *Security {
	if s == nil && other == nil {
		return nil
	}

	if other == nil {
		return s.DeepCopy()
	}

	return other.DeepCopy()
}

func (s *Security) GetSecurityContext() core.SecurityContext {
	if s == nil {
		return core.SecurityContext{}
	}

	return *s.SecurityContext
}

func (s *Security) Validate() error {
	return nil
}
