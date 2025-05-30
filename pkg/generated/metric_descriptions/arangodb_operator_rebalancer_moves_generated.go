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
	arangodbOperatorRebalancerMovesGenerated = metrics.NewDescription("arangodb_operator_rebalancer_moves_generated", "Define how many moves were generated", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorRebalancerMovesGenerated)
}

func NewArangodbOperatorRebalancerMovesGeneratedCounterFactory() metrics.FactoryCounter[ArangodbOperatorRebalancerMovesGeneratedInput] {
	return metrics.NewFactoryCounter[ArangodbOperatorRebalancerMovesGeneratedInput]()
}

func NewArangodbOperatorRebalancerMovesGeneratedInput(namespace string, name string) ArangodbOperatorRebalancerMovesGeneratedInput {
	return ArangodbOperatorRebalancerMovesGeneratedInput{
		Namespace: namespace,
		Name:      name,
	}
}

type ArangodbOperatorRebalancerMovesGeneratedInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (i ArangodbOperatorRebalancerMovesGeneratedInput) Counter(value float64) metrics.Metric {
	return ArangodbOperatorRebalancerMovesGeneratedCounter(value, i.Namespace, i.Name)
}

func (i ArangodbOperatorRebalancerMovesGeneratedInput) Desc() metrics.Description {
	return ArangodbOperatorRebalancerMovesGenerated()
}

func ArangodbOperatorRebalancerMovesGenerated() metrics.Description {
	return arangodbOperatorRebalancerMovesGenerated
}

func ArangodbOperatorRebalancerMovesGeneratedCounter(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorRebalancerMovesGenerated().Counter(value, namespace, name)
}
