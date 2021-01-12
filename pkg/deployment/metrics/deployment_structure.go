package metrics

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/prometheus/client_golang/prometheus"
)

func newDeploymentStructure(deployment *deployment) Collector {
	d := &deploymentStructure{
		deployment: deployment,
	}

	return d
}

type deploymentStructure struct {
	deployment *deployment
}

func (d deploymentStructure) Collect(metrics MetricCollector) error {
	d.deployment.deployment.Status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, member := range list {
			metrics.Collect(d.deployment.DeploymentMembers, prometheus.GaugeValue, 1, d.deployment.labels(group.AsRole(), member.ID)...)
		}

		return nil
	})
	return nil
}
