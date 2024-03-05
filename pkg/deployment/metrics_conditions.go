//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

type ConditionsMetricsMap map[api.ConditionType]bool

type ConditionsMetrics struct {
	lock sync.Mutex

	conditions ConditionsMetricsMap

	memberConditions map[string]ConditionsMetricsMap
}

func (c *ConditionsMetrics) CollectMetrics(namespace, name string, m metrics.PushMetric) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for k, v := range c.conditions {
		m.Push(metric_descriptions.ArangodbOperatorDeploymentConditionsGauge(util.BoolSwitch[float64](v, 1, 0), namespace, name, string(k)))
	}

	for member := range c.memberConditions {
		for k, v := range c.memberConditions[member] {
			m.Push(metric_descriptions.ArangodbOperatorMembersConditionsGauge(util.BoolSwitch[float64](v, 1, 0), namespace, name, member, string(k)))
		}
	}
}

func (c *ConditionsMetrics) RefreshDeployment(conditions api.ConditionList, types ...api.ConditionType) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.conditions = c.extractConditionsMap(conditions, types...)
}

func (c *ConditionsMetrics) RefreshMembers(members api.DeploymentStatusMemberElements, types ...api.ConditionType) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(members) == 0 {
		c.memberConditions = nil
		return
	}

	ret := make(map[string]ConditionsMetricsMap, len(members))

	for _, member := range members {
		ret[member.Member.ID] = c.extractConditionsMap(member.Member.Conditions, types...)
	}

	c.memberConditions = ret
}

func (c *ConditionsMetrics) extractConditionsMap(conditions api.ConditionList, types ...api.ConditionType) ConditionsMetricsMap {
	if len(types) == 0 {
		return nil
	}

	ret := make(ConditionsMetricsMap, len(types))
	for _, t := range types {
		ret[t] = conditions.IsTrue(t)
	}

	return ret
}
