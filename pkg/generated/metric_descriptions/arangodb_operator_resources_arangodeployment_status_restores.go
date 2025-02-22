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
	arangodbOperatorResourcesArangodeploymentStatusRestores = metrics.NewDescription("arangodb_operator_resources_arangodeployment_status_restores", "Counter for deployment status restored", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorResourcesArangodeploymentStatusRestores)
}

func NewArangodbOperatorResourcesArangodeploymentStatusRestoresCounterFactory() metrics.FactoryCounter[ArangodbOperatorResourcesArangodeploymentStatusRestoresInput] {
	return metrics.NewFactoryCounter[ArangodbOperatorResourcesArangodeploymentStatusRestoresInput]()
}

func NewArangodbOperatorResourcesArangodeploymentStatusRestoresInput(namespace string, name string) ArangodbOperatorResourcesArangodeploymentStatusRestoresInput {
	return ArangodbOperatorResourcesArangodeploymentStatusRestoresInput{
		Namespace: namespace,
		Name:      name,
	}
}

type ArangodbOperatorResourcesArangodeploymentStatusRestoresInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (i ArangodbOperatorResourcesArangodeploymentStatusRestoresInput) Counter(value float64) metrics.Metric {
	return ArangodbOperatorResourcesArangodeploymentStatusRestoresCounter(value, i.Namespace, i.Name)
}

func (i ArangodbOperatorResourcesArangodeploymentStatusRestoresInput) Desc() metrics.Description {
	return ArangodbOperatorResourcesArangodeploymentStatusRestores()
}

func ArangodbOperatorResourcesArangodeploymentStatusRestores() metrics.Description {
	return arangodbOperatorResourcesArangodeploymentStatusRestores
}

func ArangodbOperatorResourcesArangodeploymentStatusRestoresCounter(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorResourcesArangodeploymentStatusRestores().Counter(value, namespace, name)
}
