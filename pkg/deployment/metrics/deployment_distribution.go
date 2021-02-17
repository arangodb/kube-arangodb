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

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"github.com/prometheus/client_golang/prometheus"
)

func newDeploymentDistribution(deployment *deployment) Collector {
	d := &deploymentDistribution{
		deployment: deployment,
	}

	return d
}

type deploymentDistribution struct {
	deployment *deployment
}

type servers map[string]serverDetails

func (s *servers) update(member string, f func(m *serverDetails)) {
	if d, ok := (*s)[member]; ok {
		f(&d)
		(*s)[member] = d
	} else {
		d = serverDetails{}
		f(&d)
		(*s)[member] = d
	}
}

func (s *servers) AddCurrentShard(member string) {
	s.update(member, func(m *serverDetails) {
		m.CurrentShards++
	})
}

func (s *servers) AddPlannedShard(member string) {
	s.update(member, func(m *serverDetails) {
		m.PlannedShards++
	})
}

func (s *servers) AddLeaderShards(member string) {
	s.update(member, func(m *serverDetails) {
		m.LeaderShards++
	})
}

type serverDetails struct {
	PlannedShards int
	CurrentShards int
	LeaderShards  int
}

func (d deploymentDistribution) Collect(metrics MetricCollector) error {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	c, err := d.deployment.cache.GetAgency(ctx)
	if err != nil {
		return err
	}

	fetcher := agency.NewFetcher(c)

	plan, err := agency.GetAgencyCollections(ctx, fetcher)
	if err != nil {
		return err
	}

	current, err := agency.GetCurrentCollections(ctx, fetcher)
	if err != nil {
		return err
	}

	clusterServers := servers{}

	d.deployment.deployment.Status.Members.ForeachServerInGroups(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			clusterServers[m.ID] = serverDetails{}
		}
		return nil
	}, api.ServerGroupDBServers)

	plannedShards := map[string]map[string]bool{}

	for dbName, db := range *plan {
		for collectionID, collection := range db {
			if collection.WriteConcern != nil {
				metrics.Collect(d.deployment.DeploymentCollectionConfig, prometheus.GaugeValue, float64(*collection.WriteConcern), d.deployment.labels(dbName, collection.GetName(collectionID), "WriteConcern")...)
			}
			if collection.ReplicationFactor != nil {
				metrics.Collect(d.deployment.DeploymentCollectionConfig, prometheus.GaugeValue, float64(*collection.ReplicationFactor), d.deployment.labels(dbName, collection.GetName(collectionID), "ReplicationFactor")...)
			}

			for shardName, servers := range collection.Shards {
				for _, server := range servers {
					clusterServers.AddPlannedShard(server)

					if _, ok := plannedShards[shardName]; ok {
						if _, ok := plannedShards[shardName][server]; !ok {
							plannedShards[shardName][server] = true
						}
					} else {
						plannedShards[shardName] = map[string]bool{
							server: true,
						}
					}
				}
			}
		}
	}

	for dbName, db := range *current {
		planDB, ok := (*plan)[dbName]
		if !ok {
			continue
		}

		for collectionID, collection := range db {
			planCol, ok := planDB[collectionID]
			if !ok {
				continue
			}
			name := planCol.GetName(collectionID)

			for shardName, shard := range collection {
				if len(shard.Servers) > 0 {
					clusterServers.AddLeaderShards(shard.Servers[0])
				}

				for id, server := range shard.Servers {
					clusterServers.AddCurrentShard(server)

					var r float64 = 0
					var leader = "false"
					if id == 0 {
						leader = "true"
					}
					if p, ok := plannedShards[shardName]; ok {
						if s, ok := p[server]; ok {
							if s {
								r = 1
							}
						}
					}
					metrics.Collect(d.deployment.DeploymentShardDistribution, prometheus.GaugeValue, r, d.deployment.labels(dbName, name, shardName, server, leader)...)
				}

				if planCol.WriteConcern != nil {
					minReplicationFactor := *planCol.WriteConcern
					if len(shard.Servers) > minReplicationFactor {
						metrics.Collect(d.deployment.DeploymentShardConditions, prometheus.GaugeValue, 1, d.deployment.labels(dbName, name, shardName, "Healthy")...)
					} else if len(shard.Servers) == minReplicationFactor {
						metrics.Collect(d.deployment.DeploymentShardConditions, prometheus.GaugeValue, 1, d.deployment.labels(dbName, name, shardName, "AtMinReplicationFactor")...)
					} else {
						metrics.Collect(d.deployment.DeploymentShardConditions, prometheus.GaugeValue, 1, d.deployment.labels(dbName, name, shardName, "Offline")...)
					}
				}
			}
		}
	}

	globalDetails := serverDetails{}

	for _, details := range clusterServers {
		globalDetails.LeaderShards += details.LeaderShards
		globalDetails.CurrentShards += details.CurrentShards
		globalDetails.PlannedShards += details.PlannedShards
	}

	for member, details := range clusterServers {
		metrics.Collect(d.deployment.DeploymentServerShards, prometheus.GaugeValue, float64(details.PlannedShards), d.deployment.labels(member, "Planned")...)
		metrics.Collect(d.deployment.DeploymentServerShards, prometheus.GaugeValue, float64(details.CurrentShards), d.deployment.labels(member, "Current")...)
		metrics.Collect(d.deployment.DeploymentServerShards, prometheus.GaugeValue, float64(details.LeaderShards), d.deployment.labels(member, "Leader")...)

		metrics.Collect(d.deployment.DeploymentServerShards, prometheus.GaugeValue, float64(details.PlannedShards)/float64(globalDetails.PlannedShards), d.deployment.labels(member, "PercentagePlanned")...)
		metrics.Collect(d.deployment.DeploymentServerShards, prometheus.GaugeValue, float64(details.CurrentShards)/float64(globalDetails.CurrentShards), d.deployment.labels(member, "PercentageCurrent")...)
		metrics.Collect(d.deployment.DeploymentServerShards, prometheus.GaugeValue, float64(details.LeaderShards)/float64(globalDetails.LeaderShards), d.deployment.labels(member, "PercentageLeader")...)
	}

	metrics.Collect(d.deployment.Deployment, prometheus.GaugeValue, float64(globalDetails.PlannedShards), d.deployment.labels("Shards", "Planned")...)
	metrics.Collect(d.deployment.Deployment, prometheus.GaugeValue, float64(globalDetails.CurrentShards), d.deployment.labels("Shards", "Current")...)
	metrics.Collect(d.deployment.Deployment, prometheus.GaugeValue, float64(globalDetails.LeaderShards), d.deployment.labels("Shards", "Leader")...)

	return nil
}
