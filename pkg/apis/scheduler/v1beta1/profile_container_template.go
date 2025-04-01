//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v1beta1

import (
	core "k8s.io/api/core/v1"

	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ProfileContainerTemplate struct {
	// Containers applies values per container
	Containers schedulerContainerApi.Containers `json:"containers,omitempty"`

	// All applies generic values to all Containers
	All *schedulerContainerApi.Generic `json:"all,omitempty"`

	// Default applies generic values to default Container (first one on the list)
	Default *schedulerContainerApi.Container `json:"default,omitempty"`
}

func (p *ProfileContainerTemplate) ApplyContainers(template *core.PodTemplateSpec) error {
	if p == nil {
		return nil
	}

	return p.Containers.Apply(template)
}

func (p *ProfileContainerTemplate) ApplyGeneric(template *core.PodTemplateSpec) error {
	if p == nil {
		return nil
	}

	return p.All.Apply(template)
}

func (p *ProfileContainerTemplate) ApplyDefault(template *core.PodTemplateSpec) error {
	if p == nil {
		return nil
	}

	if template == nil {
		return errors.Errorf("Template is nil")
	}

	if len(template.Spec.Containers) == 0 {
		return errors.Errorf("Default container is missing")
	}

	var cont core.Container

	template.Spec.Containers[0].DeepCopyInto(&cont)

	if err := p.Default.Apply(template, &cont); err != nil {
		return err
	}

	template.Spec.Containers[0] = cont

	return nil
}

func (p *ProfileContainerTemplate) With(other *ProfileContainerTemplate) *ProfileContainerTemplate {
	if p == nil && other == nil {
		return nil
	}

	if p == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return p.DeepCopy()
	}

	return &ProfileContainerTemplate{
		Containers: p.Containers.With(other.Containers),
		All:        p.All.With(other.All),
		Default:    p.Default.With(other.Default),
	}
}

func (p *ProfileContainerTemplate) Validate() error {
	if p == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("containers", p.Containers.Validate()),
		shared.PrefixResourceErrors("all", p.All.Validate()),
	)
}
