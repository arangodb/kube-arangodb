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
	arangodbOperatorObjectsProcessed = metrics.NewDescription("arangodb_operator_objects_processed", "Number of the processed objects", []string{`operator_name`}, nil)

	// Global Fields
	globalArangodbOperatorObjectsProcessedCounter = NewArangodbOperatorObjectsProcessedCounterFactory()
)

func init() {
	registerDescription(arangodbOperatorObjectsProcessed)
	registerCollector(globalArangodbOperatorObjectsProcessedCounter)
}

func GlobalArangodbOperatorObjectsProcessedCounter() metrics.FactoryCounter[ArangodbOperatorObjectsProcessedInput] {
	return globalArangodbOperatorObjectsProcessedCounter
}

func NewArangodbOperatorObjectsProcessedCounterFactory() metrics.FactoryCounter[ArangodbOperatorObjectsProcessedInput] {
	return metrics.NewFactoryCounter[ArangodbOperatorObjectsProcessedInput]()
}

func NewArangodbOperatorObjectsProcessedInput(operatorName string) ArangodbOperatorObjectsProcessedInput {
	return ArangodbOperatorObjectsProcessedInput{
		OperatorName: operatorName,
	}
}

type ArangodbOperatorObjectsProcessedInput struct {
	OperatorName string `json:"operatorName"`
}

func (i ArangodbOperatorObjectsProcessedInput) Counter(value float64) metrics.Metric {
	return ArangodbOperatorObjectsProcessedCounter(value, i.OperatorName)
}

func (i ArangodbOperatorObjectsProcessedInput) Desc() metrics.Description {
	return ArangodbOperatorObjectsProcessed()
}

func ArangodbOperatorObjectsProcessed() metrics.Description {
	return arangodbOperatorObjectsProcessed
}

func ArangodbOperatorObjectsProcessedCounter(value float64, operatorName string) metrics.Metric {
	return ArangodbOperatorObjectsProcessed().Counter(value, operatorName)
}
