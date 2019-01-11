//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	//papi "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	prometheusMonitoringResourceName = "arangodb-exporter"
)

// EnsureMonitoringResources ensures that all requested external monitoring resources are created
func (r *Resources) EnsureMonitoringResources() error {
	spec := r.context.GetSpec()
	metrics := spec.Metrics

	if !metrics.IsEnabled() {
		return nil
	}

	if err := r.ensurePrometheusMonitoringResources(metrics.Prometheus); err != nil {
		return err
	}

	return nil
}

func (r *Resources) ensurePrometheusMonitoringResources(p api.MetricsPrometheusSpec) error {
	return nil
}
