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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type Counter interface {
	Metric
}

func newCounter(desc Description, value float64, labels ...string) Counter {
	return couter{
		desc:   desc,
		labels: labels,
		value:  value,
	}
}

type couter struct {
	desc Description

	labels []string

	value float64
}

func (g couter) Desc() *prometheus.Desc {
	return g.desc.Desc()
}

func (g couter) Write(metric *dto.Metric) error {
	metric.Label = g.desc.Labels(g.labels...)

	metric.Counter = &dto.Counter{
		Value: &g.value,
	}

	return nil
}
