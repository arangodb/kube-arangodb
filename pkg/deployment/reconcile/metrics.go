//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

type Metrics struct {
	Rebalancer MetricsRebalancer
}

func (m *Metrics) GetRebalancer() *MetricsRebalancer {
	if m == nil {
		return nil
	}

	return &m.Rebalancer
}

type MetricsRebalancer struct {
	enabled bool
	moves   int
	current int

	succeeded, failed int
}

func (m *MetricsRebalancer) SetEnabled(enabled bool) {
	if m == nil {
		return
	}
	m.enabled = enabled
}

func (m *MetricsRebalancer) AddMoves(moves int) {
	if m == nil {
		return
	}
	m.moves += moves
}

func (m *MetricsRebalancer) SetCurrent(current int) {
	if m == nil {
		return
	}
	m.current = current
}

func (m *MetricsRebalancer) AddFailures(i int) {
	if m == nil {
		return
	}
	m.failed += i
}

func (m *MetricsRebalancer) AddSuccesses(i int) {
	if m == nil {
		return
	}
	m.succeeded += i
}

func (r *Reconciler) CollectMetrics(m metrics.PushMetric) {
	if r.metrics.Rebalancer.enabled {
		m.Push(metric_descriptions.ArangodbOperatorRebalancerEnabled().Gauge(1, r.namespace, r.name))
		m.Push(metric_descriptions.ArangodbOperatorRebalancerMovesCurrent().Gauge(float64(r.metrics.Rebalancer.current), r.namespace, r.name))
		m.Push(metric_descriptions.ArangodbOperatorRebalancerMovesGenerated().Gauge(float64(r.metrics.Rebalancer.moves), r.namespace, r.name))
		m.Push(metric_descriptions.ArangodbOperatorRebalancerMovesSucceeded().Gauge(float64(r.metrics.Rebalancer.succeeded), r.namespace, r.name))
		m.Push(metric_descriptions.ArangodbOperatorRebalancerMovesFailed().Gauge(float64(r.metrics.Rebalancer.failed), r.namespace, r.name))
	} else {
		m.Push(metric_descriptions.ArangodbOperatorRebalancerEnabled().Gauge(0, r.namespace, r.name))
	}
}
