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

var _ interfaces.Pod[ServiceAccount] = &ServiceAccount{}

type ServiceAccount struct {

	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// AutomountServiceAccountToken indicates whether a service account token should be automatically mounted.
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty"`
}

func (s *ServiceAccount) Apply(template *core.PodTemplateSpec) error {
	if s == nil {
		return nil
	}

	c := s.DeepCopy()

	template.Spec.ServiceAccountName = c.ServiceAccountName
	if c.AutomountServiceAccountToken != nil {
		template.Spec.AutomountServiceAccountToken = c.AutomountServiceAccountToken
	}

	return nil
}

func (s *ServiceAccount) With(newResources *ServiceAccount) *ServiceAccount {
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

func (s *ServiceAccount) Validate() error {
	return nil
}
