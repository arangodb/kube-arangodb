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

var _ interfaces.Pod[Security] = &Security{}

type Security struct {
	// PodSecurityContext holds pod-level security attributes and common container settings.
	// +doc/type: core.PodSecurityContext
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	PodSecurityContext *core.PodSecurityContext `json:"podSecurityContext,omitempty"`
}

func (s *Security) Apply(template *core.PodTemplateSpec) error {
	if s == nil {
		return nil
	}

	template.Spec.SecurityContext = s.PodSecurityContext.DeepCopy()

	return nil
}

func (s *Security) GetSecurityContext() *core.PodSecurityContext {
	if s == nil {
		return nil
	}

	return s.PodSecurityContext
}

func (s *Security) With(newResources *Security) *Security {
	if s == nil && newResources == nil {
		return nil
	}

	if s == nil {
		return newResources.DeepCopy()
	}

	if newResources == nil {
		return s.DeepCopy()
	}

	return newResources.DeepCopy()
}

func (s *Security) Validate() error {
	return nil
}
