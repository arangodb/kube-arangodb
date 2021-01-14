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
	typedApi "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/deployment/v1"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Deployments interface {
	prometheus.Collector
}

func NewDeployments(metrics MetricDefinition, deploymentInterface typedApi.ArangoDeploymentInterface, cli kubernetes.Interface) Deployments {
	d := &deployments{
		deploymentInterface: deploymentInterface,
		cli:                 cli,
		MetricDefinition:    metrics,
	}

	return d
}

type deployments struct {
	deploymentInterface typedApi.ArangoDeploymentInterface
	cli                 kubernetes.Interface

	MetricDefinition
}

func (d deployments) Collect(metrics chan<- prometheus.Metric) {
	collector := NewMetricsCollector(metrics)

	println("COLLECTING DATA")

	depls, err := d.deploymentInterface.List(v1.ListOptions{})

	defer d.Error.ApplyMetrics(collector)
	if err != nil {
		d.Error.Add()
		collector.Collect(d.DeploymentCount, prometheus.GaugeValue, 0)
		return
	}

	collector.Collect(d.DeploymentCount, prometheus.GaugeValue, float64(len(depls.Items)))

	for _, depl := range depls.Items {
		NewDeployment(d.MetricDefinition, d.cli, &depl).Collect(metrics)
	}
}
