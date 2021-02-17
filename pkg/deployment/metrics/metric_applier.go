//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

type MetricApplier interface {
	ApplyMetrics(a MetricCollector)
}

type MetricCollector interface {
	Collect(m Metric, valueType prometheus.ValueType, value float64, labelValues ...string) MetricCollector

	CollectInstance(a MetricApplier) MetricCollector
}

type applier struct {
	metrics chan<- prometheus.Metric
}

func (a *applier) CollectInstance(app MetricApplier) MetricCollector {
	app.ApplyMetrics(a)
	return a
}

func (a *applier) Collect(m Metric, valueType prometheus.ValueType, value float64, labelValues ...string) MetricCollector {
	if metric, err := prometheus.NewConstMetric(m.Desc(), valueType, value, labelValues...); err == nil {
		a.metrics <- metric
	}

	return a
}

func NewMetricsCollector(metrics chan<- prometheus.Metric) MetricCollector {
	return &applier{metrics: metrics}
}
