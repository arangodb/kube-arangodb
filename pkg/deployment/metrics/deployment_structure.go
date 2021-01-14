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
