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
