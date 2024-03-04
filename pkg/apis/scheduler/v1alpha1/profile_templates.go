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

package v1alpha1

import (
	"sort"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ProfileTemplates []ProfileTemplate

func (p ProfileTemplates) Sort() ProfileTemplates {
	sort.Slice(p, func(i, j int) bool {
		if a, b := util.WithDefault(p[i].Priority), util.WithDefault(p[j].Priority); a != b {
			return a < b
		}

		return false
	})

	return p
}

func (p ProfileTemplates) Render() (*core.PodTemplateSpec, error) {
	var pod core.PodTemplateSpec

	// Apply Pod Spec
	for id := range p {
		if err := p[id].Pod.Apply(&pod); err != nil {
			return nil, errors.Wrapf(err, "Error while rendering Pod for %d", id)
		}
	}
	// Apply Containers Spec
	for id := range p {
		if err := p[id].Container.ApplyContainers(&pod); err != nil {
			return nil, errors.Wrapf(err, "Error while rendering Pod for %d", id)
		}
	}
	// Apply Generic Containers Spec
	for id := range p {
		if err := p[id].Container.ApplyGeneric(&pod); err != nil {
			return nil, errors.Wrapf(err, "Error while rendering Pod for %d", id)
		}
	}

	return pod.DeepCopy(), nil
}
