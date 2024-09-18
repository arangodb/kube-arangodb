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
	arangodbOperatorKubernetesClientRequests = metrics.NewDescription("arangodb_operator_kubernetes_client_requests", "Number of Kubernetes Client requests", []string{`component`, `verb`}, nil)
)

func init() {
	registerDescription(arangodbOperatorKubernetesClientRequests)
}

func NewArangodbOperatorKubernetesClientRequestsCounterFactory() metrics.FactoryCounter[ArangodbOperatorKubernetesClientRequestsInput] {
	return metrics.NewFactoryCounter[ArangodbOperatorKubernetesClientRequestsInput]()
}

func NewArangodbOperatorKubernetesClientRequestsInput(component string, verb string) ArangodbOperatorKubernetesClientRequestsInput {
	return ArangodbOperatorKubernetesClientRequestsInput{
		Component: component,
		Verb:      verb,
	}
}

type ArangodbOperatorKubernetesClientRequestsInput struct {
	Component string `json:"component"`
	Verb      string `json:"verb"`
}

func (i ArangodbOperatorKubernetesClientRequestsInput) Counter(value float64) metrics.Metric {
	return ArangodbOperatorKubernetesClientRequestsCounter(value, i.Component, i.Verb)
}

func (i ArangodbOperatorKubernetesClientRequestsInput) Desc() metrics.Description {
	return ArangodbOperatorKubernetesClientRequests()
}

func ArangodbOperatorKubernetesClientRequests() metrics.Description {
	return arangodbOperatorKubernetesClientRequests
}

func ArangodbOperatorKubernetesClientRequestsCounter(value float64, component string, verb string) metrics.Metric {
	return ArangodbOperatorKubernetesClientRequests().Counter(value, component, verb)
}
