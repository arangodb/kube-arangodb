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
	arangodbResourcesDeploymentConfigMapDuration = metrics.NewDescription("arangodb_resources_deployment_config_map_duration", "Duration of inspected ConfigMaps by Deployment in seconds", []string{`deployment`}, nil)

	// Global Fields
	globalArangodbResourcesDeploymentConfigMapDurationGauge = NewArangodbResourcesDeploymentConfigMapDurationGaugeFactory()
)

func init() {
	registerDescription(arangodbResourcesDeploymentConfigMapDuration)
	registerCollector(globalArangodbResourcesDeploymentConfigMapDurationGauge)
}

func GlobalArangodbResourcesDeploymentConfigMapDurationGauge() metrics.FactoryGauge[ArangodbResourcesDeploymentConfigMapDurationInput] {
	return globalArangodbResourcesDeploymentConfigMapDurationGauge
}

func NewArangodbResourcesDeploymentConfigMapDurationGaugeFactory() metrics.FactoryGauge[ArangodbResourcesDeploymentConfigMapDurationInput] {
	return metrics.NewFactoryGauge[ArangodbResourcesDeploymentConfigMapDurationInput]()
}

func NewArangodbResourcesDeploymentConfigMapDurationInput(deployment string) ArangodbResourcesDeploymentConfigMapDurationInput {
	return ArangodbResourcesDeploymentConfigMapDurationInput{
		Deployment: deployment,
	}
}

type ArangodbResourcesDeploymentConfigMapDurationInput struct {
	Deployment string `json:"deployment"`
}

func (i ArangodbResourcesDeploymentConfigMapDurationInput) Gauge(value float64) metrics.Metric {
	return ArangodbResourcesDeploymentConfigMapDurationGauge(value, i.Deployment)
}

func (i ArangodbResourcesDeploymentConfigMapDurationInput) Desc() metrics.Description {
	return ArangodbResourcesDeploymentConfigMapDuration()
}

func ArangodbResourcesDeploymentConfigMapDuration() metrics.Description {
	return arangodbResourcesDeploymentConfigMapDuration
}

func ArangodbResourcesDeploymentConfigMapDurationGauge(value float64, deployment string) metrics.Metric {
	return ArangodbResourcesDeploymentConfigMapDuration().Gauge(value, deployment)
}
