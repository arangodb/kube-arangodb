//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package chart

import (
	"context"

	"helm.sh/helm/v3/pkg/chart"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

var logger = logging.Global().RegisterAndGetLogger("platform-chart-operator", logging.Info)

type handler struct {
	client     arangoClientSet.Interface
	kubeClient kubernetes.Interface

	eventRecorder event.RecorderInstance

	operator operator.Operator
}

func (h *handler) Name() string {
	return Kind()
}

func (h *handler) Handle(ctx context.Context, item operation.Item) error {
	// Get Backup object. It also covers NotFound case
	object, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.client.PlatformV1beta1().ArangoPlatformCharts(item.Namespace).Get, item.Name, meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	status := object.Status.DeepCopy()

	changed, reconcileErr := operator.HandleP3WithStop(ctx, item, object, status, h.handle)
	if reconcileErr != nil && !operator.IsReconcile(reconcileErr) {
		logger.Err(reconcileErr).Warn("Fail for %s %s/%s",
			item.Kind,
			item.Namespace,
			item.Name)

		return reconcileErr
	}

	if !changed {
		return reconcileErr
	}

	logger.Debug("Updating %s %s/%s",
		item.Kind,
		item.Namespace,
		item.Name)

	if _, err := operator.WithArangoPlatformChartUpdateStatusInterfaceRetry(context.Background(), h.client.PlatformV1beta1().ArangoPlatformCharts(object.GetNamespace()), object, *status, meta.UpdateOptions{}); err != nil {
		return err
	}

	return reconcileErr
}

func (h *handler) handle(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformChart, status *platformApi.ArangoPlatformChartStatus) (bool, error) {
	return operator.HandleP3WithCondition(ctx, &status.Conditions, platformApi.ReadyCondition, item, extension, status, h.HandleSpecValidity, h.HandleSpecData)
}

func (h *handler) HandleSpecValidity(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformChart, status *platformApi.ArangoPlatformChartStatus) (bool, error) {
	if err := extension.Spec.Validate(); err != nil {
		// We have received an error in the spec!

		logger.Err(err).Warn("Invalid Spec on %s", item.String())

		if status.Conditions.Update(platformApi.SpecValidCondition, false, "Spec is invalid", "Spec is invalid") {
			return true, operator.Stop("Invalid spec")
		}
		return false, operator.Stop("Invalid spec")
	}

	if status.Conditions.Update(platformApi.SpecValidCondition, true, "Spec is valid", "Spec is valid") {
		return true, nil
	}

	return false, nil
}

func (h *handler) HandleSpecData(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformChart, status *platformApi.ArangoPlatformChartStatus) (bool, error) {
	checksum := extension.Spec.Checksum()

	if status.Info != nil {
		if status.Info.Checksum != checksum {
			status.Info = nil
			return true, operator.Reconcile("Spec changed")
		}

		if !status.Info.Valid {
			return false, operator.Stop("Invalid Chart")
		}

		// All fine
		return false, nil
	}

	chart, err := helm.Chart(extension.Spec.Definition).Get()
	if err != nil {
		status.Info = &platformApi.ChartStatusInfo{
			Definition: extension.Spec.Definition,
			Checksum:   checksum,
			Overrides:  extension.Spec.Overrides,
			Valid:      false,
			Message:    "Chart is invalid",
		}

		return true, operator.Reconcile("Spec changed")
	}

	if chart.Chart().Name() != extension.GetName() {
		status.Info = &platformApi.ChartStatusInfo{
			Definition: extension.Spec.Definition,
			Checksum:   checksum,
			Overrides:  extension.Spec.Overrides,
			Valid:      false,
			Message:    "Chart Name mismatch",
		}

		return true, operator.Reconcile("Spec changed")
	}

	platform, err := chart.Platform()
	if err != nil {
		status.Info = &platformApi.ChartStatusInfo{
			Definition: extension.Spec.Definition,
			Checksum:   checksum,
			Overrides:  extension.Spec.Overrides,
			Valid:      false,
			Message:    "Chart is invalid: Unable to get platform details",
		}

		return true, operator.Reconcile("Spec changed")
	}

	status.Info = &platformApi.ChartStatusInfo{
		Definition: extension.Spec.Definition,
		Checksum:   checksum,
		Overrides:  extension.Spec.Overrides,
		Valid:      true,
		Details:    chartInfoExtract(chart.Chart(), platform),
	}

	h.eventRecorder.Normal(extension, "Chart Accepted", "Chart Accepted with checksum: %s", checksum)

	return true, operator.Reconcile("Spec changed")
}

func (h *handler) CanBeHandled(item operation.Item) bool {
	return item.Group == Group() &&
		utilConstants.Version(Version()).IsCompatible(utilConstants.Version(item.Version)) &&
		item.Kind == Kind()
}

func chartInfoExtract(chart *chart.Chart, platform *helm.Platform) *platformApi.ChartDetails {
	if chart == nil || chart.Metadata == nil {
		return nil
	}

	r := &platformApi.ChartDetails{
		Name:    chart.Name(),
		Version: chart.Metadata.Version,
	}

	if platform != nil {
		var c platformApi.ChartDetailsPlatform

		if len(platform.Requirements) > 0 {
			c.Requirements = make(platformApi.ChartDetailsPlatformRequirements, len(platform.Requirements))

			for k, v := range platform.Requirements {
				c.Requirements[k] = platformApi.ChartDetailsPlatformVersionConstrain(v)
			}
		}

		r.Platform = &c
	}

	return r
}
