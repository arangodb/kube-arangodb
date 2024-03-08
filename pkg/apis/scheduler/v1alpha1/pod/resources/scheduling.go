//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/interfaces"
	"github.com/arangodb/kube-arangodb/pkg/util"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tolerations"
)

type Tolerations []core.Toleration

var _ interfaces.Pod[Scheduling] = &Scheduling{}

type Scheduling struct {
	// NodeSelector is a selector that must be true for the workload to fit on a node.
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Affinity defines scheduling constraints for workload
	// +doc/type: core.Affinity
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity
	Affinity *core.Affinity `json:"affinity,omitempty"`

	// Tolerations defines tolerations
	// +doc/type: []core.Toleration
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	Tolerations Tolerations `json:"tolerations,omitempty"`

	// SchedulerName specifies, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +doc/default: ""
	SchedulerName *string `json:"schedulerName,omitempty"`
}

func (s *Scheduling) Apply(template *core.PodTemplateSpec) error {
	if s == nil {
		return nil
	}

	if len(s.NodeSelector) == 0 {
		template.Spec.NodeSelector = nil
	} else {
		template.Spec.NodeSelector = map[string]string{}
		for k, v := range s.NodeSelector {
			template.Spec.NodeSelector[k] = v
		}
	}

	if s.Affinity != nil {
		if s.Affinity.NodeAffinity != nil || s.Affinity.PodAffinity != nil || s.Affinity.PodAntiAffinity != nil {
			template.Spec.Affinity = s.Affinity.DeepCopy()
		}
	}

	template.Spec.Tolerations = tolerations.AddTolerationsIfNotFound(nil, s.Tolerations.DeepCopy()...)

	template.Spec.SchedulerName = util.WithDefault(s.SchedulerName)

	return nil
}

func (s *Scheduling) GetNodeSelector() map[string]string {
	if s != nil {
		return s.NodeSelector
	}

	return nil
}

func (s *Scheduling) GetSchedulerName() string {
	if s != nil && s.SchedulerName != nil {
		return *s.SchedulerName
	}

	return ""
}

func (s *Scheduling) GetAffinity() *core.Affinity {
	if s != nil {
		return s.Affinity
	}

	return nil
}

func (s *Scheduling) GetTolerations() Tolerations {
	if s != nil {
		return s.Tolerations
	}

	return nil
}

func (s *Scheduling) With(other *Scheduling) *Scheduling {
	if s == nil && other == nil {
		return nil
	}

	if other == nil {
		return s.DeepCopy()
	}

	if s == nil {
		return other.DeepCopy()
	}

	current := s.DeepCopy()
	new := other.DeepCopy()

	// NodeSelector
	if len(current.NodeSelector) == 0 {
		current.NodeSelector = new.NodeSelector
	} else if len(new.NodeSelector) > 0 {
		for k, v := range new.NodeSelector {
			current.NodeSelector[k] = v
		}
	}

	// SchedulerName
	if new.SchedulerName != nil {
		current.SchedulerName = new.SchedulerName
	}

	// Tolerations
	current.Tolerations = tolerations.AddTolerationsIfNotFound(new.Tolerations, other.Tolerations...)

	// Affinity
	current.Affinity = kresources.Merge(current.Affinity, new.Affinity)

	// return

	if current.Affinity == nil &&
		current.SchedulerName == nil &&
		len(current.Tolerations) == 0 &&
		len(current.NodeSelector) == 0 {
		return nil
	}

	return current
}

func (s *Scheduling) Validate() error {
	return nil
}
