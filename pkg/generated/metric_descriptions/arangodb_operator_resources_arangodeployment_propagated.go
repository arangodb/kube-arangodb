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
	arangodbOperatorResourcesArangodeploymentPropagated = metrics.NewDescription("arangodb_operator_resources_arangodeployment_propagated", "Defines if ArangoDeployment Spec is propagated", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorResourcesArangodeploymentPropagated)
}

func NewArangodbOperatorResourcesArangodeploymentPropagatedGaugeFactory() metrics.FactoryGauge[ArangodbOperatorResourcesArangodeploymentPropagatedInput] {
	return metrics.NewFactoryGauge[ArangodbOperatorResourcesArangodeploymentPropagatedInput]()
}

func NewArangodbOperatorResourcesArangodeploymentPropagatedInput(namespace string, name string) ArangodbOperatorResourcesArangodeploymentPropagatedInput {
	return ArangodbOperatorResourcesArangodeploymentPropagatedInput{
		Namespace: namespace,
		Name:      name,
	}
}

type ArangodbOperatorResourcesArangodeploymentPropagatedInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (i ArangodbOperatorResourcesArangodeploymentPropagatedInput) Gauge(value float64) metrics.Metric {
	return ArangodbOperatorResourcesArangodeploymentPropagatedGauge(value, i.Namespace, i.Name)
}

func (i ArangodbOperatorResourcesArangodeploymentPropagatedInput) Desc() metrics.Description {
	return ArangodbOperatorResourcesArangodeploymentPropagated()
}

func ArangodbOperatorResourcesArangodeploymentPropagated() metrics.Description {
	return arangodbOperatorResourcesArangodeploymentPropagated
}

func ArangodbOperatorResourcesArangodeploymentPropagatedGauge(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorResourcesArangodeploymentPropagated().Gauge(value, namespace, name)
}
