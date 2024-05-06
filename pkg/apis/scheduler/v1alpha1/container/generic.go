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

	schedulerContainerResourcesApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/container/resources"
	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/interfaces"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

var _ interfaces.Pod[Generic] = &Generic{}

type Generic struct {
	// Environments keeps the environment variables for Container
	*schedulerContainerResourcesApiv1alpha1.Environments `json:",inline"`

	// VolumeMounts define volume mounts assigned to the Container
	*schedulerContainerResourcesApiv1alpha1.VolumeMounts `json:",inline"`
}

func (g *Generic) Apply(template *core.PodTemplateSpec) error {
	if g == nil {
		return nil
	}

	for id := range template.Spec.Containers {
		if err := shared.WithErrors(
			g.Environments.Apply(template, &template.Spec.Containers[id]),
			g.VolumeMounts.Apply(template, &template.Spec.Containers[id]),
		); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generic) GetEnvironments() *schedulerContainerResourcesApiv1alpha1.Environments {
	if g == nil || g.Environments == nil {
		return nil
	}

	return g.Environments
}

func (g *Generic) GetVolumeMounts() *schedulerContainerResourcesApiv1alpha1.VolumeMounts {
	if g == nil || g.VolumeMounts == nil {
		return nil
	}

	return g.VolumeMounts
}

func (g *Generic) With(other *Generic) *Generic {
	if g == nil && other == nil {
		return nil
	}

	if g == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return g.DeepCopy()
	}

	return &Generic{
		Environments: g.Environments.With(other.Environments),
		VolumeMounts: g.VolumeMounts.With(other.VolumeMounts),
	}
}

func (g *Generic) Validate() error {
	if g == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("containerEnvironments", g.Environments.Validate()),
		shared.PrefixResourceErrors("volumeMounts", g.VolumeMounts.Validate()),
	)
}
