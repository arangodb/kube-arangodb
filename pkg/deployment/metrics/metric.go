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

type Counter interface {
	Add()

	MetricApplier
	MetricDesc
}

type counter struct {
	Metric
	labels []string

	count int64
}

func (c *counter) Add() {
	c.count++
}

func (c *counter) ApplyMetrics(a MetricCollector) {
	a.Collect(c, prometheus.CounterValue, float64(c.count), c.labels...)
}

type MetricDesc interface {
	Desc() *prometheus.Desc
}

type Metric interface {
	MetricDesc

	NewCounter(labels ...string) Counter
}

func NewMetricDescription(fqName, help string, variableLabels []string) Metric {
	return metricDescription{prometheus.NewDesc(fqName, help, variableLabels, nil)}
}

type metricDescription struct {
	Description *prometheus.Desc
}

func (m metricDescription) NewCounter(labels ...string) Counter {
	return &counter{
		Metric: m,
		count:  0,
		labels: labels,
	}
}

func (m metricDescription) Desc() *prometheus.Desc {
	return m.Description
}
