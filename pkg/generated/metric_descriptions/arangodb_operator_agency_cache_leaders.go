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
	arangodbOperatorAgencyCacheLeaders = metrics.NewDescription("arangodb_operator_agency_cache_leaders", "Determines agency leader vote count", []string{`namespace`, `name`, `agent`}, nil)
)

func init() {
	registerDescription(arangodbOperatorAgencyCacheLeaders)
}

func NewArangodbOperatorAgencyCacheLeadersGaugeFactory() metrics.FactoryGauge[ArangodbOperatorAgencyCacheLeadersInput] {
	return metrics.NewFactoryGauge[ArangodbOperatorAgencyCacheLeadersInput]()
}

func NewArangodbOperatorAgencyCacheLeadersInput(namespace string, name string, agent string) ArangodbOperatorAgencyCacheLeadersInput {
	return ArangodbOperatorAgencyCacheLeadersInput{
		Namespace: namespace,
		Name:      name,
		Agent:     agent,
	}
}

type ArangodbOperatorAgencyCacheLeadersInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Agent     string `json:"agent"`
}

func (i ArangodbOperatorAgencyCacheLeadersInput) Gauge(value float64) metrics.Metric {
	return ArangodbOperatorAgencyCacheLeadersGauge(value, i.Namespace, i.Name, i.Agent)
}

func (i ArangodbOperatorAgencyCacheLeadersInput) Desc() metrics.Description {
	return ArangodbOperatorAgencyCacheLeaders()
}

func ArangodbOperatorAgencyCacheLeaders() metrics.Description {
	return arangodbOperatorAgencyCacheLeaders
}

func ArangodbOperatorAgencyCacheLeadersGauge(value float64, namespace string, name string, agent string) metrics.Metric {
	return ArangodbOperatorAgencyCacheLeaders().Gauge(value, namespace, name, agent)
}
