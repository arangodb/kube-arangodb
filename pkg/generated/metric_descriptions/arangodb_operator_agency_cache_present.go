//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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
	arangodbOperatorAgencyCachePresent = metrics.NewDescription("arangodb_operator_agency_cache_present", "Determines if local agency cache is present", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorAgencyCachePresent)
}

func NewArangodbOperatorAgencyCachePresentGaugeFactory() metrics.FactoryGauge[ArangodbOperatorAgencyCachePresentInput] {
	return metrics.NewFactoryGauge[ArangodbOperatorAgencyCachePresentInput]()
}

func NewArangodbOperatorAgencyCachePresentInput(namespace string, name string) ArangodbOperatorAgencyCachePresentInput {
	return ArangodbOperatorAgencyCachePresentInput{
		Namespace: namespace,
		Name:      name,
	}
}

type ArangodbOperatorAgencyCachePresentInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (i ArangodbOperatorAgencyCachePresentInput) Gauge(value float64) metrics.Metric {
	return ArangodbOperatorAgencyCachePresentGauge(value, i.Namespace, i.Name)
}

func (i ArangodbOperatorAgencyCachePresentInput) Desc() metrics.Description {
	return ArangodbOperatorAgencyCachePresent()
}

func ArangodbOperatorAgencyCachePresent() metrics.Description {
	return arangodbOperatorAgencyCachePresent
}

func ArangodbOperatorAgencyCachePresentGauge(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorAgencyCachePresent().Gauge(value, namespace, name)
}
