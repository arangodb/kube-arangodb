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
	arangodbOperatorDeploymentConditions = metrics.NewDescription("arangodb_operator_deployment_conditions", "Representation of the ArangoDeployment condition state (true/false)", []string{`namespace`, `name`, `condition`}, nil)
)

func init() {
	registerDescription(arangodbOperatorDeploymentConditions)
}

func ArangodbOperatorDeploymentConditions() metrics.Description {
	return arangodbOperatorDeploymentConditions
}

func ArangodbOperatorDeploymentConditionsGauge(value float64, namespace string, name string, condition string) metrics.Metric {
	return ArangodbOperatorDeploymentConditions().Gauge(value, namespace, name, condition)
}
