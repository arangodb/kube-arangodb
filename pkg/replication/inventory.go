//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package replication

import (
	"sync"

	"github.com/arangodb/kube-arangodb/pkg/metrics/collector"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

func init() {
	localInventory = inventory{
		deploymentReplications: map[string]map[string]*DeploymentReplication{},
	}
	collector.GetCollector().RegisterMetric(&localInventory)
}

var localInventory inventory

type inventory struct {
	lock                   sync.Mutex
	deploymentReplications map[string]map[string]*DeploymentReplication
}

func (i *inventory) CollectMetrics(in metrics.PushMetric) {
	for _, drs := range i.deploymentReplications {
		for _, dr := range drs {
			dr.CollectMetrics(in)
		}
	}
}

func (i *inventory) Add(dr *DeploymentReplication) {
	i.lock.Lock()
	defer i.lock.Unlock()

	name, namespace := dr.apiObject.GetName(), dr.apiObject.GetNamespace()
	if _, ok := i.deploymentReplications[namespace]; !ok {
		i.deploymentReplications[namespace] = map[string]*DeploymentReplication{}
	}
	i.deploymentReplications[namespace][name] = dr
}
