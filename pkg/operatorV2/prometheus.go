//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

package operator

import (
	"github.com/prometheus/client_golang/prometheus"
)

type prometheusMetrics struct {
	operator *operator
}

func newCollector(operator *operator) *prometheusMetrics {
	return &prometheusMetrics{
		operator: operator,
	}
}

func (p *prometheusMetrics) connectors() []prometheus.Collector {
	return []prometheus.Collector{}
}

func (p *prometheusMetrics) Describe(r chan<- *prometheus.Desc) {
	for _, c := range p.connectors() {
		c.Describe(r)
	}

	for _, h := range p.operator.handlers {
		if collector, ok := h.(prometheus.Collector); ok {
			collector.Describe(r)
		}
	}
}

func (p *prometheusMetrics) Collect(r chan<- prometheus.Metric) {
	for _, c := range p.connectors() {
		c.Collect(r)
	}

	for _, h := range p.operator.handlers {
		if collector, ok := h.(prometheus.Collector); ok {
			collector.Collect(r)
		}
	}
}
