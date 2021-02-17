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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
)

type Deployment interface {
	prometheus.Collector
}

func NewDeployment(metrics MetricDefinition, cli kubernetes.Interface, d *api.ArangoDeployment) Deployment {
	depl := &deployment{
		MetricDefinition: metrics,
		deployment:       d,
	}

	depl.cache = client.NewClientCache(depl.getDeployment, conn.NewFactory(client.NewAuth(cli, depl.getDeployment), client.NewConfig(depl.getDeployment)))

	depl.distribution = newDeploymentDistribution(depl)
	depl.structure = newDeploymentStructure(depl)

	return depl
}

type deployment struct {
	MetricDefinition

	deployment *api.ArangoDeployment

	cache client.Cache

	distribution Collector
	structure    Collector
}

func (d *deployment) getDeployment() *api.ArangoDeployment {
	return d.deployment
}

func (d *deployment) Collect(metrics chan<- prometheus.Metric) {
	collector := NewMetricsCollector(metrics)

	if err := d.distribution.Collect(collector); err != nil {
		d.Error.Add()
	}

	if err := d.structure.Collect(collector); err != nil {
		d.Error.Add()
	}
}

func (d *deployment) labels(labels ...string) []string {
	return append([]string{
		d.deployment.GetName(),
		d.deployment.GetNamespace(),
	}, labels...)
}
