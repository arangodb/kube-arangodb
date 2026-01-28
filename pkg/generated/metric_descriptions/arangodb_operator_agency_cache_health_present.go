//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package metric_descriptions

import (
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

var (
	arangodbOperatorAgencyCacheHealthPresent = metrics.NewDescription("arangodb_operator_agency_cache_health_present", "Determines if local agency cache health is present", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorAgencyCacheHealthPresent)
}

func NewArangodbOperatorAgencyCacheHealthPresentGaugeFactory() metrics.FactoryGauge[ArangodbOperatorAgencyCacheHealthPresentInput] {
	return metrics.NewFactoryGauge[ArangodbOperatorAgencyCacheHealthPresentInput]()
}

func NewArangodbOperatorAgencyCacheHealthPresentInput(namespace string, name string) ArangodbOperatorAgencyCacheHealthPresentInput {
	return ArangodbOperatorAgencyCacheHealthPresentInput{
		Namespace: namespace,
		Name:      name,
	}
}

type ArangodbOperatorAgencyCacheHealthPresentInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (i ArangodbOperatorAgencyCacheHealthPresentInput) Gauge(value float64) metrics.Metric {
	return ArangodbOperatorAgencyCacheHealthPresentGauge(value, i.Namespace, i.Name)
}

func (i ArangodbOperatorAgencyCacheHealthPresentInput) Desc() metrics.Description {
	return ArangodbOperatorAgencyCacheHealthPresent()
}

func ArangodbOperatorAgencyCacheHealthPresent() metrics.Description {
	return arangodbOperatorAgencyCacheHealthPresent
}

func ArangodbOperatorAgencyCacheHealthPresentGauge(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorAgencyCacheHealthPresent().Gauge(value, namespace, name)
}
