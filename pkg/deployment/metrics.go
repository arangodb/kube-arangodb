//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/util/metrics"
)

type Metrics struct {
	Agency struct {
		Errors  uint64
		Fetches uint64
		Index   uint64
	}

	Errors struct {
		DeploymentValidationErrors, DeploymentImmutableErrors, StatusRestores uint64
	}

	ArangodbOperatorEngineOpsAlerts int

	Deployment struct {
		Accepted, UpToDate, Propagated bool
	}
}

func (d *Deployment) CollectMetrics(m metrics.PushMetric) {
	m.Push(metric_descriptions.ArangodbOperatorAgencyErrorsCounter(float64(d.metrics.Agency.Errors), d.namespace, d.name))
	m.Push(metric_descriptions.ArangodbOperatorAgencyFetchesCounter(float64(d.metrics.Agency.Fetches), d.namespace, d.name))
	m.Push(metric_descriptions.ArangodbOperatorAgencyIndexGauge(float64(d.metrics.Agency.Index), d.namespace, d.name))

	m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentValidationErrorsCounter(float64(d.metrics.Errors.DeploymentValidationErrors), d.namespace, d.name))
	m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentImmutableErrorsCounter(float64(d.metrics.Errors.DeploymentImmutableErrors), d.namespace, d.name))
	m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentStatusRestoresCounter(float64(d.metrics.Errors.StatusRestores), d.namespace, d.name))

	if d.metrics.Deployment.Accepted {
		m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentAcceptedGauge(1, d.namespace, d.name))
	} else {
		m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentAcceptedGauge(0, d.namespace, d.name))
	}
	if d.metrics.Deployment.UpToDate {
		m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentUptodateGauge(1, d.namespace, d.name))
	} else {
		m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentUptodateGauge(0, d.namespace, d.name))
	}
	if d.metrics.Deployment.Propagated {
		m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentPropagatedGauge(1, d.namespace, d.name))
	} else {
		m.Push(metric_descriptions.ArangodbOperatorResourcesArangodeploymentPropagatedGauge(0, d.namespace, d.name))
	}

	if c := d.agencyCache; c != nil {
		m.Push(metric_descriptions.ArangodbOperatorAgencyCachePresentGauge(1, d.namespace, d.name))
		if h, ok := c.Health(); ok {
			m.Push(metric_descriptions.ArangodbOperatorAgencyCacheHealthPresentGauge(1, d.namespace, d.name))

			h.CollectMetrics(m)
		} else {
			m.Push(metric_descriptions.ArangodbOperatorAgencyCacheHealthPresentGauge(0, d.namespace, d.name))
		}
	} else {
		m.Push(metric_descriptions.ArangodbOperatorAgencyCachePresentGauge(0, d.namespace, d.name))
	}

	// Reconcile
	if c := d.reconciler; c != nil {
		c.CollectMetrics(m)
	}

	// Resources
	if r := d.resources; r != nil {
		r.CollectMetrics(m)
	}
}
