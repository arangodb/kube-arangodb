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

package container

import (
	core "k8s.io/api/core/v1"

	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/interfaces"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

var _ interfaces.Pod[Generic] = &Generic{}

type Generic struct {
	// Environments keeps the environment variables for Container
	*schedulerContainerResourcesApi.Environments `json:",inline"`
}

func (c *Generic) Apply(template *core.PodTemplateSpec) error {
	if c == nil {
		return nil
	}

	for id := range template.Spec.Containers {
		if err := shared.WithErrors(
			c.Environments.Apply(template, &template.Spec.Containers[id]),
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *Generic) GetEnvironments() *schedulerContainerResourcesApi.Environments {
	if c == nil || c.Environments == nil {
		return nil
	}

	return c.Environments
}

func (c *Generic) With(other *Generic) *Generic {
	if c == nil && other == nil {
		return nil
	}

	if c == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return c.DeepCopy()
	}

	return &Generic{
		Environments: c.Environments.With(other.Environments),
	}
}

func (c *Generic) Validate() error {
	if c == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("containerEnvironments", c.Environments.Validate()),
	)
}
