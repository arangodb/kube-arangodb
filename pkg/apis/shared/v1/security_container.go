//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

import core "k8s.io/api/core/v1"

type SecurityContainer struct {
	// PodSecurityContext holds pod-level security attributes and common container settings.
	// +doc/type: core.SecurityContext
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	SecurityContext *core.SecurityContext `json:"securityContext,omitempty"`
}

func (s *SecurityContainer) With(other *SecurityContainer) *SecurityContainer {
	if s == nil && other == nil {
		return nil
	}

	if other == nil {
		return s.DeepCopy()
	}

	// TODO: Add fine graned merge

	return other.DeepCopy()
}

func (s *SecurityContainer) GetSecurityContext() *core.SecurityContext {
	if s == nil {
		return nil
	}

	return s.SecurityContext
}

func (s *SecurityContainer) Validate() error {
	return nil
}
