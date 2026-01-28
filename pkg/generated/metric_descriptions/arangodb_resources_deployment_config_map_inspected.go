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
	arangodbResourcesDeploymentConfigMapInspected = metrics.NewDescription("arangodb_resources_deployment_config_map_inspected", "Number of inspected ConfigMaps by Deployment", []string{`deployment`}, nil)

	// Global Fields
	globalArangodbResourcesDeploymentConfigMapInspectedCounter = NewArangodbResourcesDeploymentConfigMapInspectedCounterFactory()
)

func init() {
	registerDescription(arangodbResourcesDeploymentConfigMapInspected)
	registerCollector(globalArangodbResourcesDeploymentConfigMapInspectedCounter)
}

func GlobalArangodbResourcesDeploymentConfigMapInspectedCounter() metrics.FactoryCounter[ArangodbResourcesDeploymentConfigMapInspectedInput] {
	return globalArangodbResourcesDeploymentConfigMapInspectedCounter
}

func NewArangodbResourcesDeploymentConfigMapInspectedCounterFactory() metrics.FactoryCounter[ArangodbResourcesDeploymentConfigMapInspectedInput] {
	return metrics.NewFactoryCounter[ArangodbResourcesDeploymentConfigMapInspectedInput]()
}

func NewArangodbResourcesDeploymentConfigMapInspectedInput(deployment string) ArangodbResourcesDeploymentConfigMapInspectedInput {
	return ArangodbResourcesDeploymentConfigMapInspectedInput{
		Deployment: deployment,
	}
}

type ArangodbResourcesDeploymentConfigMapInspectedInput struct {
	Deployment string `json:"deployment"`
}

func (i ArangodbResourcesDeploymentConfigMapInspectedInput) Counter(value float64) metrics.Metric {
	return ArangodbResourcesDeploymentConfigMapInspectedCounter(value, i.Deployment)
}

func (i ArangodbResourcesDeploymentConfigMapInspectedInput) Desc() metrics.Description {
	return ArangodbResourcesDeploymentConfigMapInspected()
}

func ArangodbResourcesDeploymentConfigMapInspected() metrics.Description {
	return arangodbResourcesDeploymentConfigMapInspected
}

func ArangodbResourcesDeploymentConfigMapInspectedCounter(value float64, deployment string) metrics.Metric {
	return ArangodbResourcesDeploymentConfigMapInspected().Counter(value, deployment)
}
