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
	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

type Metrics struct {
	DeploymentReplication struct {
		Active, Failed bool
	}
}

func (dr *DeploymentReplication) CollectMetrics(m metrics.PushMetric) {
	name, namespace := dr.apiObject.GetName(), dr.apiObject.GetNamespace()

	m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentreplicationActiveGauge(
		util.BoolSwitch[float64](dr.metrics.DeploymentReplication.Active, 1, 0), namespace, name),
	)
	m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentreplicationFailedGauge(
		util.BoolSwitch[float64](dr.metrics.DeploymentReplication.Failed, 1, 0), namespace, name),
	)
}
