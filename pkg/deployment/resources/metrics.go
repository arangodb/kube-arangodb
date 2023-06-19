//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

const (
	// Component name for metrics of this package
	metricsComponent = "deployment_resources"
)

type Metrics struct {
	lock sync.Mutex

	Members map[string]MetricMember
}

func (m *Metrics) IncMemberContainerRestarts(id, container, reason string, code int32) {
	if m == nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.Members == nil {
		m.Members = map[string]MetricMember{}
	}

	v := m.Members[id]

	if v.ContainerRestarts == nil {
		v.ContainerRestarts = map[string]MetricMemberRestarts{}
	}

	cr := v.ContainerRestarts[container]

	if cr == nil {
		cr = MetricMemberRestarts{}
	}

	cd := cr[code]

	if cd == nil {
		cd = MetricMemberRestartReason{}
	}

	cd[reason]++

	cr[code] = cd

	v.ContainerRestarts[container] = cr

	m.Members[id] = v
}

func (m *Metrics) IncMemberInitContainerRestarts(id, container, reason string, code int32) {
	if m == nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.Members == nil {
		m.Members = map[string]MetricMember{}
	}

	v := m.Members[id]

	if v.InitContainerRestarts == nil {
		v.InitContainerRestarts = map[string]MetricMemberRestarts{}
	}

	cr := v.InitContainerRestarts[container]

	if cr == nil {
		cr = MetricMemberRestarts{}
	}

	cd := cr[code]

	if cd == nil {
		cd = MetricMemberRestartReason{}
	}

	cd[reason]++

	cr[code] = cd

	v.InitContainerRestarts[container] = cr

	m.Members[id] = v
}

func (m *Metrics) IncMemberEphemeralContainerRestarts(id, container, reason string, code int32) {
	if m == nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.Members == nil {
		m.Members = map[string]MetricMember{}
	}

	v := m.Members[id]

	if v.EphemeralContainerRestarts == nil {
		v.EphemeralContainerRestarts = map[string]MetricMemberRestarts{}
	}

	cr := v.EphemeralContainerRestarts[container]

	if cr == nil {
		cr = MetricMemberRestarts{}
	}

	cd := cr[code]

	if cd == nil {
		cd = MetricMemberRestartReason{}
	}

	cd[reason]++

	cr[code] = cd

	v.EphemeralContainerRestarts[container] = cr

	m.Members[id] = v
}

type MetricMember struct {
	ContainerRestarts          MetricContainerRestarts
	InitContainerRestarts      MetricContainerRestarts
	EphemeralContainerRestarts MetricContainerRestarts
}

type MetricContainerRestarts map[string]MetricMemberRestarts

type MetricMemberRestarts map[int32]MetricMemberRestartReason

type MetricMemberRestartReason map[string]uint64

func (r *Resources) CollectMetrics(m metrics.PushMetric) {
	for member, info := range r.metrics.Members {
		// Containers
		for container, restarts := range info.ContainerRestarts {
			for code, reasons := range restarts {
				for reason, count := range reasons {
					m.Push(metric_descriptions.ArangodbOperatorMembersUnexpectedContainerExitCodesCounter(float64(count), r.namespace, r.name, member, container, "container", fmt.Sprintf("%d", code), reason))
				}
			}
		}
		// InitContainers
		for container, restarts := range info.InitContainerRestarts {
			for code, reasons := range restarts {
				for reason, count := range reasons {
					m.Push(metric_descriptions.ArangodbOperatorMembersUnexpectedContainerExitCodesCounter(float64(count), r.namespace, r.name, member, container, "initContainer", fmt.Sprintf("%d", code), reason))
				}
			}
		}
		// EphemeralContainers
		for container, restarts := range info.EphemeralContainerRestarts {
			for code, reasons := range restarts {
				for reason, count := range reasons {
					m.Push(metric_descriptions.ArangodbOperatorMembersUnexpectedContainerExitCodesCounter(float64(count), r.namespace, r.name, member, container, "ephemeralContainer", fmt.Sprintf("%d", code), reason))
				}
			}
		}
	}
}
