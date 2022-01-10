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

type PushDescription interface {
	Push(desc ...Description) PushDescription
}

type PushMetric interface {
	Push(desc ...Metric) PushMetric
}

type pushDescription struct {
	out chan<- *prometheus.Desc
}

func (p pushDescription) Push(desc ...Description) PushDescription {
	for _, q := range desc {
		p.out <- q.Desc()
	}
	return p
}

type pushMetric struct {
	out chan<- prometheus.Metric
}

func (p pushMetric) Push(desc ...Metric) PushMetric {
	for _, q := range desc {
		p.out <- q
	}
	return p
}

func NewPushDescription(out chan<- *prometheus.Desc) PushDescription {
	return pushDescription{
		out: out,
	}
}

func NewPushMetric(out chan<- prometheus.Metric) PushMetric {
	return pushMetric{
		out: out,
	}
}
