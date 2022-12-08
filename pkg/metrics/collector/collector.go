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

package collector

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

type Collector interface {
	RegisterMetric(m metrics.MCollector)
	RegisterDescription(m metrics.DCollector)

	SetFilter(filter metrics.MetricPushFilter)
}

func init() {
	prometheus.MustRegister(collectorObject)
}

func GetCollector() Collector {
	return collectorObject
}

var collectorObject = &collector{}

type collector struct {
	lock sync.Mutex

	filter metrics.MetricPushFilter

	metrics      []metrics.MCollector
	descriptions []metrics.DCollector
}

func (p *collector) SetFilter(filter metrics.MetricPushFilter) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.filter = filter
}

func (p *collector) RegisterDescription(m metrics.DCollector) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.descriptions = append(p.descriptions, m)
}

func (p *collector) RegisterMetric(m metrics.MCollector) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.metrics = append(p.metrics, m)
}

func (p *collector) Describe(descs chan<- *prometheus.Desc) {
	p.lock.Lock()
	defer p.lock.Unlock()

	out := metrics.NewPushDescription(descs)

	for id := range p.descriptions {
		p.descriptions[id].CollectDescriptions(out)
	}
}

func (p *collector) Collect(c chan<- prometheus.Metric) {
	p.lock.Lock()
	defer p.lock.Unlock()

	out := metrics.NewPushMetric(c)

	if f := p.filter; f != nil {
		out = metrics.NewMetricsPushFilter(out, f)
	}

	for id := range p.metrics {
		p.metrics[id].CollectMetrics(out)
	}
}
