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
	arangodbOperatorEngineOpsAlerts = metrics.NewDescription("arangodb_operator_engine_ops_alerts", "Counter for actions which requires ops attention", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorEngineOpsAlerts)
}

func NewArangodbOperatorEngineOpsAlertsCounterFactory() metrics.FactoryCounter[ArangodbOperatorEngineOpsAlertsInput] {
	return metrics.NewFactoryCounter[ArangodbOperatorEngineOpsAlertsInput]()
}

func NewArangodbOperatorEngineOpsAlertsInput(namespace string, name string) ArangodbOperatorEngineOpsAlertsInput {
	return ArangodbOperatorEngineOpsAlertsInput{
		Namespace: namespace,
		Name:      name,
	}
}

type ArangodbOperatorEngineOpsAlertsInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (i ArangodbOperatorEngineOpsAlertsInput) Counter(value float64) metrics.Metric {
	return ArangodbOperatorEngineOpsAlertsCounter(value, i.Namespace, i.Name)
}

func (i ArangodbOperatorEngineOpsAlertsInput) Desc() metrics.Description {
	return ArangodbOperatorEngineOpsAlerts()
}

func ArangodbOperatorEngineOpsAlerts() metrics.Description {
	return arangodbOperatorEngineOpsAlerts
}

func ArangodbOperatorEngineOpsAlertsCounter(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorEngineOpsAlerts().Counter(value, namespace, name)
}
