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

package assertion

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/metrics/collector"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

func init() {
	collector.GetCollector().RegisterMetric(metricsObject)
}

type metricsObjectType struct {
	metrics map[Key]int
	lock    sync.Mutex
}

var (
	metricsObject = &metricsObjectType{
		metrics: map[Key]int{},
	}
)

func (m *metricsObjectType) incKeyMetric(key Key) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.metrics[key]++
}

func (m *metricsObjectType) CollectMetrics(in metrics.PushMetric) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for key, invokes := range m.metrics {
		in.Push(
			metric_descriptions.ArangodbOperatorEngineAssertionsCounter(float64(invokes), string(key)),
		)
	}
}
