//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

import "github.com/arangodb/kube-arangodb/pkg/util/metrics"

var (
	arangodbOperatorResourcesArangodeploymentValidationErrors = metrics.NewDescription("arangodb_operator_resources_arangodeployment_validation_errors", "Counter for deployment validation errors", []string{`namespace`, `name`}, nil)
)

func init() {
	registerDescription(arangodbOperatorResourcesArangodeploymentValidationErrors)
}

func ArangodbOperatorResourcesArangodeploymentValidationErrors() metrics.Description {
	return arangodbOperatorResourcesArangodeploymentValidationErrors
}

func ArangodbOperatorResourcesArangodeploymentValidationErrorsCounter(value float64, namespace string, name string) metrics.Metric {
	return ArangodbOperatorResourcesArangodeploymentValidationErrors().Counter(value, namespace, name)
}
