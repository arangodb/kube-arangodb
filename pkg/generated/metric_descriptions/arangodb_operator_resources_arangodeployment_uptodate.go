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
	arangodbOperatorResourcesArangodeploymentUptodate = metrics.NewDescription("arangodb_operator_resources_arangodeployment_uptodate", "Defines if ArangoDeployment is uptodate", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorResourcesArangodeploymentUptodate)
}

func NewArangodbOperatorResourcesArangodeploymentUptodateGaugeFactory() metrics.FactoryGauge[ArangodbOperatorResourcesArangodeploymentUptodateInput] {
	return metrics.NewFactoryGauge[ArangodbOperatorResourcesArangodeploymentUptodateInput]()
}

func NewArangodbOperatorResourcesArangodeploymentUptodateInput(namespace string, name string) ArangodbOperatorResourcesArangodeploymentUptodateInput {
	return ArangodbOperatorResourcesArangodeploymentUptodateInput{
		Namespace: namespace,
		Name:      name,
	}
}

type ArangodbOperatorResourcesArangodeploymentUptodateInput struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (i ArangodbOperatorResourcesArangodeploymentUptodateInput) Gauge(value float64) metrics.Metric {
	return ArangodbOperatorResourcesArangodeploymentUptodateGauge(value, i.Namespace, i.Name)
}

func (i ArangodbOperatorResourcesArangodeploymentUptodateInput) Desc() metrics.Description {
	return ArangodbOperatorResourcesArangodeploymentUptodate()
}

func ArangodbOperatorResourcesArangodeploymentUptodate() metrics.Description {
	return arangodbOperatorResourcesArangodeploymentUptodate
}

func ArangodbOperatorResourcesArangodeploymentUptodateGauge(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorResourcesArangodeploymentUptodate().Gauge(value, namespace, name)
}
