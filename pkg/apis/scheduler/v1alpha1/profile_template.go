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
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/pod"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ProfileTemplate struct {
	Priority *int `json:"priority,omitempty"`

	Pod *schedulerPodApi.Pod `json:"pod,omitempty"`

	Container *ProfileContainerTemplate `json:"container,omitempty"`
}

func (p *ProfileTemplate) GetPod() *schedulerPodApi.Pod {
	if p == nil || p.Pod == nil {
		return nil
	}

	return p.Pod
}

func (p *ProfileTemplate) GetContainer() *ProfileContainerTemplate {
	if p == nil || p.Container == nil {
		return nil
	}

	return p.Container
}

func (p *ProfileTemplate) GetPriority() int {
	if p == nil || p.Priority == nil {
		return 0
	}

	return *p.Priority
}

func (p *ProfileTemplate) With(other *ProfileTemplate) *ProfileTemplate {
	if p == nil && other == nil {
		return nil
	}

	if p == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return p.DeepCopy()
	}

	return &ProfileTemplate{
		Priority:  util.NewType(max(p.GetPriority(), other.GetPriority())),
		Pod:       p.Pod.With(other.Pod),
		Container: p.Container.With(other.Container),
	}
}

func (p *ProfileTemplate) Validate() error {
	if p == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("pod", p.Pod.Validate()),
		shared.PrefixResourceErrors("container", p.Container.Validate()),
	)
}
