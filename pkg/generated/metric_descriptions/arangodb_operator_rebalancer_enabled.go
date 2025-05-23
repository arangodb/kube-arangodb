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
	arangodbOperatorRebalancerEnabled = metrics.NewDescription("arangodb_operator_rebalancer_enabled", "Determines if rebalancer is enabled", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorRebalancerEnabled)
}

func NewArangodbOperatorRebalancerEnabledGaugeFactory() metrics.FactoryGauge[ArangodbOperatorRebalancerEnabledInput] {
	return metrics.NewFactoryGauge[ArangodbOperatorRebalancerEnabledInput]()
}

func NewArangodbOperatorRebalancerEnabledInput(namespace string, name string) ArangodbOperatorRebalancerEnabledInput {
	return ArangodbOperatorRebalancerEnabledInput{
		Namespace: namespace,
		Name:      name,
	}
}

type ArangodbOperatorRebalancerEnabledInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (i ArangodbOperatorRebalancerEnabledInput) Gauge(value float64) metrics.Metric {
	return ArangodbOperatorRebalancerEnabledGauge(value, i.Namespace, i.Name)
}

func (i ArangodbOperatorRebalancerEnabledInput) Desc() metrics.Description {
	return ArangodbOperatorRebalancerEnabled()
}

func ArangodbOperatorRebalancerEnabled() metrics.Description {
	return arangodbOperatorRebalancerEnabled
}

func ArangodbOperatorRebalancerEnabledGauge(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorRebalancerEnabled().Gauge(value, namespace, name)
}
