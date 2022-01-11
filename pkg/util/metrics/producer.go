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

package metrics

import "github.com/prometheus/client_golang/prometheus"

type Producer interface {
	Gauge(value float64, labels ...string) Producer
}

type producer struct {
	d   *description
	out chan<- prometheus.Metric
}

func (p *producer) Gauge(value float64, labels ...string) Producer {
	p.out <- newGauge(p.d, value, labels...)
	return p
}

func newProducer(out chan<- prometheus.Metric, d *description) Producer {
	return &producer{
		d:   d,
		out: out,
	}
}
