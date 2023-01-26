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

package inspector

import (
	"reflect"
	"sync"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/metrics/collector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

func init() {
	collector.GetCollector().RegisterMetric(clientMetricsInstance)
}

var (
	clientMetricsInstance = &clientMetrics{}
)

type clientMetricsFields struct {
	calls  int
	errors int
}

type clientMetrics struct {
	lock sync.Mutex

	metrics map[definitions.Component]map[definitions.Verb]clientMetricsFields
}

func (c *clientMetrics) reset() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.metrics = nil
}

func (c *clientMetrics) CollectMetrics(in metrics.PushMetric) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for component, verbs := range c.metrics {
		for verb, fields := range verbs {
			in.Push(
				metric_descriptions.ArangodbOperatorKubernetesClientRequestsCounter(float64(fields.calls), string(component), string(verb)),
				metric_descriptions.ArangodbOperatorKubernetesClientRequestErrorsCounter(float64(fields.errors), string(component), string(verb)),
			)
		}
	}
}

func (c *clientMetrics) ObjectRequest(definition definitions.Component, verb definitions.Verb, object meta.Object, err error) {
	if object == nil || (reflect.ValueOf(object).Kind() == reflect.Ptr && reflect.ValueOf(object).IsNil()) {
		return
	}
	c.Request(definition, verb, object.GetName(), err)
}

func (c *clientMetrics) Request(definition definitions.Component, verb definitions.Verb, name string, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.metrics == nil {
		c.metrics = map[definitions.Component]map[definitions.Verb]clientMetricsFields{}
	}

	if _, ok := c.metrics[definition]; !ok {
		c.metrics[definition] = map[definitions.Verb]clientMetricsFields{}
	}

	if _, ok := c.metrics[definition][verb]; !ok {
		c.metrics[definition][verb] = clientMetricsFields{}
	}

	f := c.metrics[definition][verb]

	f.calls++

	if err != nil {
		f.errors++
	}

	c.metrics[definition][verb] = f

	// Logging

	log := clientLogger.Str("name", name).Str("kind", string(definition)).Str("verb", string(verb)).Err(err)

	if err == nil {
		call := log.Debug

		switch verb {
		case definitions.Get:
			call = log.Trace
		case definitions.Update, definitions.UpdateStatus, definitions.Patch:
			call = log.Debug
		case definitions.Create, definitions.Delete:
			call = log.Info
		case definitions.ForceDelete:
			call = log.Warn
		}

		call("Kubernetes request has been send")
	} else {
		log.Warn("Kubernetes request failed")
	}
}
