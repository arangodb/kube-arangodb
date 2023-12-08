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

import (
	core "k8s.io/api/core/v1"
)

type SchedulingTolerations []core.Toleration

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
	Tolerations SchedulingTolerations `json:"tolerations,omitempty"`

	// SchedulerName specifies, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +doc/default: ""
	SchedulerName *string `json:"schedulerName,omitempty"`
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

func (s *Scheduling) GetTolerations() SchedulingTolerations {
	if s != nil {
		return s.Tolerations
	}

	return nil
}

func (s *Scheduling) Validate() error {
	return nil
}
