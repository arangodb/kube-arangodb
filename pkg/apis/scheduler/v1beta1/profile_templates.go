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
	"sort"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ProfileTemplates []*ProfileTemplate

func (p ProfileTemplates) Sort() ProfileTemplates {
	sort.Slice(p, func(i, j int) bool {
		if a, b := p[i].GetPriority(), p[j].GetPriority(); a != b {
			return a > b
		}

		return false
	})

	return p
}

func (p ProfileTemplates) Merge() *ProfileTemplate {
	var z *ProfileTemplate

	for id := len(p) - 1; id >= 0; id-- {
		z = z.With(p[id])
	}

	return z
}

func (p ProfileTemplates) RenderOnTemplate(pod *core.PodTemplateSpec) error {
	t := p.Merge()

	// Apply  ArangoSchedulerPod Spec
	if err := t.GetPod().Apply(pod); err != nil {
		return errors.Wrapf(err, "Error while rendering  ArangoSchedulerPod")
	}

	// Apply Generic Containers Spec
	if err := t.GetContainer().ApplyGeneric(pod); err != nil {
		return errors.Wrapf(err, "Error while rendering  ArangoSchedulerPod")
	}

	// Apply Containers Spec
	if err := t.GetContainer().ApplyContainers(pod); err != nil {
		return errors.Wrapf(err, "Error while rendering  ArangoSchedulerPod")
	}

	// Apply Default Containers Spec
	if err := t.GetContainer().ApplyDefault(pod); err != nil {
		return errors.Wrapf(err, "Error while rendering  ArangoSchedulerPod")
	}

	return nil
}
