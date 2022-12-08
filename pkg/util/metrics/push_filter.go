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

type MetricPushFilter func(m Metric) bool

func NegateMetricPushFilter(in MetricPushFilter) MetricPushFilter {
	return func(m Metric) bool {
		return !in(m)
	}
}

func MergeMetricPushFilter(filters ...MetricPushFilter) MetricPushFilter {
	return func(m Metric) bool {
		for _, f := range filters {
			if f == nil {
				continue
			}
			if !f(m) {
				return false
			}
		}

		return true
	}
}

type metricPushFilter struct {
	filter MetricPushFilter

	out PushMetric
}

func (m metricPushFilter) Push(desc ...Metric) PushMetric {
	for id := range desc {
		if m.filter(desc[id]) {
			m.out.Push(desc[id])
			continue
		}
	}

	return m
}

func NewMetricsPushFilter(out PushMetric, filter MetricPushFilter) PushMetric {
	return &metricPushFilter{
		filter: filter,
		out:    out,
	}
}
