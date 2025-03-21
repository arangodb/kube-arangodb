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
	arangodbOperatorRebalancerMovesCurrent = metrics.NewDescription("arangodb_operator_rebalancer_moves_current", "Define how many moves are currently in progress", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorRebalancerMovesCurrent)
}

func NewArangodbOperatorRebalancerMovesCurrentGaugeFactory() metrics.FactoryGauge[ArangodbOperatorRebalancerMovesCurrentInput] {
	return metrics.NewFactoryGauge[ArangodbOperatorRebalancerMovesCurrentInput]()
}

func NewArangodbOperatorRebalancerMovesCurrentInput(namespace string, name string) ArangodbOperatorRebalancerMovesCurrentInput {
	return ArangodbOperatorRebalancerMovesCurrentInput{
		Namespace: namespace,
		Name:      name,
	}
}

type ArangodbOperatorRebalancerMovesCurrentInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (i ArangodbOperatorRebalancerMovesCurrentInput) Gauge(value float64) metrics.Metric {
	return ArangodbOperatorRebalancerMovesCurrentGauge(value, i.Namespace, i.Name)
}

func (i ArangodbOperatorRebalancerMovesCurrentInput) Desc() metrics.Description {
	return ArangodbOperatorRebalancerMovesCurrent()
}

func ArangodbOperatorRebalancerMovesCurrent() metrics.Description {
	return arangodbOperatorRebalancerMovesCurrent
}

func ArangodbOperatorRebalancerMovesCurrentGauge(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorRebalancerMovesCurrent().Gauge(value, namespace, name)
}
