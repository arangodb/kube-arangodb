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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Description interface {
	Desc() *prometheus.Desc
	Labels(labels ...string) []*dto.LabelPair

	Collect(out chan<- prometheus.Metric, collect func(p Producer) error)

	Gauge(value float64, labels ...string) Metric
	Counter(value float64, labels ...string) Metric
}

type description struct {
	variableLabels []string
	constLabels    prometheus.Labels
	desc           *prometheus.Desc
}

func (d *description) Counter(value float64, labels ...string) Metric {
	return newCounter(d, value, labels...)
}

func (d *description) Gauge(value float64, labels ...string) Metric {
	return newGauge(d, value, labels...)
}

func (d *description) Collect(out chan<- prometheus.Metric, collect func(p Producer) error) {
	if err := collect(newProducer(out, d)); err != nil {
		out <- newError(d, err)
	}
}

func (d *description) Desc() *prometheus.Desc {
	return d.desc
}

func (d *description) Labels(labels ...string) []*dto.LabelPair {
	var l []*dto.LabelPair

	for k, v := range d.constLabels {
		var z dto.LabelPair

		z.Name = util.NewType[string](k)
		z.Value = util.NewType[string](v)
		l = append(l, &z)
	}

	for id := range labels {
		if id >= len(d.variableLabels) {
			break
		}

		var z dto.LabelPair

		z.Name = util.NewType[string](d.variableLabels[id])
		z.Value = util.NewType[string](labels[id])
		l = append(l, &z)
	}

	return l
}

func NewDescription(fqName, help string, variableLabels []string, constLabels prometheus.Labels) Description {
	return &description{
		variableLabels: variableLabels,
		constLabels:    constLabels,
		desc:           prometheus.NewDesc(fqName, help, variableLabels, constLabels),
	}
}
