//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package metrics

import "github.com/prometheus/client_golang/prometheus"

func NewMetricDefinition() MetricDefinition {
	return MetricDefinition{
		Error:                       NewMetricDescription("arango_operator_deployment_scrape_errors", "Deployment Scrape Errors", nil).NewCounter(),
		DeploymentCount:             NewMetricDescription("arango_operator_deployment_count", "Count of deployments managed by operator", nil),
		DeploymentCollectionConfig:  NewMetricDescription("arango_operator_deployment_collection_config", "Shard Configuration", coreLabels("database", "collection", "option")),
		DeploymentShardDistribution: NewMetricDescription("arango_operator_deployment_shard_distribution", "Shard Distribution", coreLabels("database", "collection", "shard", "server", "leader")),
		DeploymentShardConditions:   NewMetricDescription("arango_operator_deployment_shard_conditions", "Shard Conditions", coreLabels("database", "collection", "shard", "condition")),

		DeploymentServerShards: NewMetricDescription("arango_operator_deployment_server_shards", "Server Shard Distribution", coreLabels("server", "type")),

		DeploymentShards: NewMetricDescription("arango_operator_deployment_shards", "Server Shard Statistics", coreLabels("shard", "type")),

		Deployment: NewMetricDescription("arango_operator_deployment", "Deployment Statistics", coreLabels("kind", "type")),

		DeploymentMembers: NewMetricDescription("arango_operator_deployment_members", "", coreLabels("member_role", "id")),
	}
}

func coreLabels(labels ...string) []string {
	return append([]string{
		"deployment",
		"namespace",
	}, labels...)
}

type MetricDefinition struct {
	Error Counter

	DeploymentCount Metric

	DeploymentCollectionConfig  Metric
	DeploymentShardDistribution Metric
	DeploymentShardConditions   Metric

	DeploymentServerShards Metric
	Deployment             Metric

	DeploymentMembers Metric
	DeploymentShards  Metric
}

func (m MetricDefinition) Describe(descs chan<- *prometheus.Desc) {
	collector := NewDesc(descs)

	collector.Collect(m.Error)
	collector.Collect(m.DeploymentCount)
	collector.Collect(m.DeploymentCollectionConfig)
	collector.Collect(m.DeploymentShardDistribution)
	collector.Collect(m.DeploymentShardConditions)
	collector.Collect(m.DeploymentServerShards)
	collector.Collect(m.DeploymentMembers)
	collector.Collect(m.DeploymentShards)
	collector.Collect(m.Deployment)
}
