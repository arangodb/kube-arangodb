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
