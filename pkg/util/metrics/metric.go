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

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type Metric interface {
	prometheus.Metric
}

type Gauge interface {
	Metric
}

type Error interface {
	Metric
}

type errorRet struct {
	desc Description
	err  error
}

func (e errorRet) Desc() *prometheus.Desc {
	return e.desc.Desc()
}

func (e errorRet) Write(metric *dto.Metric) error {
	return e.err
}

func newError(desc Description, value error) Error {
	return errorRet{
		desc: desc,
		err:  value,
	}
}

func newGauge(desc Description, value float64, labels ...string) Gauge {
	return gauge{
		desc:   desc,
		labels: labels,
		value:  value,
	}
}

type gauge struct {
	desc Description

	labels []string

	value float64
}

func (g gauge) Desc() *prometheus.Desc {
	return g.desc.Desc()
}

func (g gauge) Write(metric *dto.Metric) error {
	metric.Label = g.desc.Labels(g.labels...)

	metric.Gauge = &dto.Gauge{
		Value: &g.value,
	}

	return nil
}
