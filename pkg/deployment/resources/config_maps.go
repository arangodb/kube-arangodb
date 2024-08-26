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
	"fmt"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	configMapsV1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/configmap/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
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

	if features.Gateway().Enabled() && spec.IsGatewayEnabled() {
		counterMetric.Inc()
		if err := reconcileRequired.WithError(r.ensureGatewayConfig(ctx, cachedStatus, configMaps)); err != nil {
			return errors.Section(err, "Gateway ConfigMap")
		}
	}
	return reconcileRequired.Reconcile(ctx)
}

func (r *Resources) ensureGatewayConfig(ctx context.Context, cachedStatus inspectorInterface.Inspector, configMaps configMapsV1.ModInterface) error {
	deploymentName := r.context.GetAPIObject().GetName()
	configMapName := GetGatewayConfigMapName(deploymentName)

	if _, exists := cachedStatus.ConfigMap().V1().GetSimple(configMapName); !exists {
		// Find serving service (single/crdn)
		spec := r.context.GetSpec()
		svcServingName := fmt.Sprintf("%s-%s", deploymentName, spec.Mode.Get().ServingGroup().AsRole())

		svc, svcExist := cachedStatus.Service().V1().GetSimple(svcServingName)
		if !svcExist {
			return errors.Errorf("Service %s not found", svcServingName)
		}

		gatewayCfgYaml, err := RenderGatewayConfigYAML(svc.Spec.ClusterIP)
		if err != nil {
			return errors.WithStack(errors.Wrapf(err, "Failed to render gateway config"))
		}
		cm := &core.ConfigMap{
			ObjectMeta: meta.ObjectMeta{
				Name: configMapName,
			},
			Data: map[string]string{
				GatewayConfigFileName:     string(gatewayCfgYaml),
				GatewayConfigChecksumName: util.SHA256(gatewayCfgYaml),
			},
		}

		owner := r.context.GetAPIObject().AsOwner()

		err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return k8sutil.CreateConfigMap(ctxChild, configMaps, cm, &owner)
		})
		if kerrors.IsAlreadyExists(err) {
			// CM added while we tried it also
			return nil
		} else if err != nil {
			// Failed to create
			return errors.WithStack(err)
		}

		return errors.Reconcile()
	}
	return nil
}
