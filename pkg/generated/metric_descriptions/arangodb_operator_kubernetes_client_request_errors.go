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
	arangodbOperatorKubernetesClientRequestErrors = metrics.NewDescription("arangodb_operator_kubernetes_client_request_errors", "Number of Kubernetes Client request errors", []string{`component`, `verb`}, nil)
)

func init() {
	registerDescription(arangodbOperatorKubernetesClientRequestErrors)
}

func NewArangodbOperatorKubernetesClientRequestErrorsCounterFactory() metrics.FactoryCounter[ArangodbOperatorKubernetesClientRequestErrorsInput] {
	return metrics.NewFactoryCounter[ArangodbOperatorKubernetesClientRequestErrorsInput]()
}

func NewArangodbOperatorKubernetesClientRequestErrorsInput(component string, verb string) ArangodbOperatorKubernetesClientRequestErrorsInput {
	return ArangodbOperatorKubernetesClientRequestErrorsInput{
		Component: component,
		Verb:      verb,
	}
}

type ArangodbOperatorKubernetesClientRequestErrorsInput struct {
	Component string `json:"component"`
	Verb      string `json:"verb"`
}

func (i ArangodbOperatorKubernetesClientRequestErrorsInput) Counter(value float64) metrics.Metric {
	return ArangodbOperatorKubernetesClientRequestErrorsCounter(value, i.Component, i.Verb)
}

func (i ArangodbOperatorKubernetesClientRequestErrorsInput) Desc() metrics.Description {
	return ArangodbOperatorKubernetesClientRequestErrors()
}

func ArangodbOperatorKubernetesClientRequestErrors() metrics.Description {
	return arangodbOperatorKubernetesClientRequestErrors
}

func ArangodbOperatorKubernetesClientRequestErrorsCounter(value float64, component string, verb string) metrics.Metric {
	return ArangodbOperatorKubernetesClientRequestErrors().Counter(value, component, verb)
}
