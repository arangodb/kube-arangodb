//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

package deployment

import (
	"sync"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Component name for metrics of this package
	metricsComponent = "deployment"
)

func init() {
	localInventory = inventory{
		deployments:                    map[string]map[string]*Deployment{},
		deploymentsMetric:              metrics.NewDescription("arangodb_operator_deployments", "Number of active deployments", []string{"namespace", "deployment"}, nil),
		deploymentMetricsMembersMetric: metrics.NewDescription("arango_operator_deployment_members", "List of members", []string{"namespace", "deployment", "role", "id"}, nil),
		deploymentAgencyStateMetric:    metrics.NewDescription("arango_operator_deployment_agency_state", "Reachability of agency", []string{"namespace", "deployment"}, nil),
		deploymentShardLeadersMetric:   metrics.NewDescription("arango_operator_deployment_shard_leaders", "Deployment leader shards distribution", []string{"namespace", "deployment", "database", "collection", "shard", "server"}, nil),
		deploymentShardsMetric:         metrics.NewDescription("arango_operator_deployment_shards", "Deployment shards distribution", []string{"namespace", "deployment", "database", "collection", "shard", "server"}, nil),
	}

	prometheus.MustRegister(&localInventory)
}

var localInventory inventory

var _ prometheus.Collector = &inventory{}

type inventory struct {
	lock        sync.Mutex
	deployments map[string]map[string]*Deployment

	deploymentsMetric, deploymentMetricsMembersMetric, deploymentAgencyStateMetric, deploymentShardsMetric, deploymentShardLeadersMetric metrics.Description
}

func (i *inventory) Describe(descs chan<- *prometheus.Desc) {
	i.lock.Lock()
	defer i.lock.Unlock()

	metrics.NewPushDescription(descs).Push(i.deploymentsMetric, i.deploymentMetricsMembersMetric, i.deploymentAgencyStateMetric, i.deploymentShardLeadersMetric, i.deploymentShardsMetric)
}

func (i *inventory) Collect(m chan<- prometheus.Metric) {
	i.lock.Lock()
	defer i.lock.Unlock()

	p := metrics.NewPushMetric(m)
	for _, deployments := range i.deployments {
		for _, deployment := range deployments {
			p.Push(i.deploymentsMetric.Gauge(1, deployment.GetNamespace(), deployment.GetName()))

			spec := deployment.GetSpec()
			status, _ := deployment.GetStatus()

			for _, member := range status.Members.AsList() {
				p.Push(i.deploymentMetricsMembersMetric.Gauge(1, deployment.GetNamespace(), deployment.GetName(), member.Group.AsRole(), member.Member.ID))
			}

			if spec.Mode.Get().HasAgents() {
				agency, agencyOk := deployment.GetAgencyCache()
				if !agencyOk {
					p.Push(i.deploymentAgencyStateMetric.Gauge(0, deployment.GetNamespace(), deployment.GetName()))
					continue
				}

				p.Push(i.deploymentAgencyStateMetric.Gauge(1, deployment.GetNamespace(), deployment.GetName()))

				if spec.Mode.Get() == api.DeploymentModeCluster {
					for db, collections := range agency.Current.Collections {
						for collection, shards := range collections {
							for shard, details := range shards {
								for id, server := range details.Servers {
									name := "UNKNOWN"
									if _, ok := agency.Plan.Collections[db]; ok {
										if _, ok := agency.Plan.Collections[db][collection]; ok {
											name = agency.Plan.Collections[db][collection].GetName(name)
										}
									}

									m := []string{
										deployment.GetNamespace(),
										deployment.GetName(),
										db,
										name,
										shard,
										server,
									}

									if id == 0 {
										p.Push(i.deploymentShardLeadersMetric.Gauge(1, m...))
									}
									p.Push(i.deploymentShardsMetric.Gauge(1, m...))
								}
							}
						}
					}
				}
			}
		}
	}
}

func (i *inventory) Add(d *Deployment) {
	i.lock.Lock()
	defer i.lock.Unlock()

	name, namespace := d.GetName(), d.GetNamespace()

	if _, ok := i.deployments[namespace]; !ok {
		i.deployments[namespace] = map[string]*Deployment{}
	}

	i.deployments[namespace][name] = d
}
