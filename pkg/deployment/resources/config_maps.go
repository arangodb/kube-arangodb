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

package resources

import (
	"context"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

var (
	inspectedConfigMapsCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_config_maps", "Number of ConfigMaps inspections per deployment", metrics.DeploymentName)
	inspectConfigMapsDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_config_maps_duration", "Amount of time taken by a single inspection of all ConfigMaps for a deployment (in sec)", metrics.DeploymentName)
)

// EnsureConfigMaps creates all ConfigMaps needed to run the given deployment
func (r *Resources) EnsureConfigMaps(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	start := time.Now()
	spec := r.context.GetSpec()
	configMaps := cachedStatus.ConfigMapsModInterface().V1()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()

	defer metrics.SetDuration(inspectConfigMapsDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedConfigMapsCounters.WithLabelValues(deploymentName)

	reconcileRequired := k8sutil.NewReconcile(cachedStatus)

	if features.IsGatewayEnabled(spec) {
		counterMetric.Inc()
		if err := reconcileRequired.WithError(r.ensureGatewayConfig(ctx, cachedStatus, configMaps)); err != nil {
			return errors.Section(err, "Gateway ConfigMap")
		}
		if err := reconcileRequired.WithError(r.ensureMemberConfig(ctx, cachedStatus, configMaps)); err != nil {
			return errors.Section(err, "Member ConfigMap")
		}
	}
	return reconcileRequired.Reconcile(ctx)
}
