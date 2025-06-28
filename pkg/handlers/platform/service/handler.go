//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package service

import (
	"context"
	"time"

	"helm.sh/helm/v3/pkg/action"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/platform"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
)

var logger = logging.Global().RegisterAndGetLogger("platform-service-operator", logging.Info)

type handler struct {
	client     arangoClientSet.Interface
	kubeClient kubernetes.Interface

	eventRecorder event.RecorderInstance

	operator operator.Operator

	helm helm.Client
}

func (h *handler) Name() string {
	return Kind()
}

func (h *handler) Handle(ctx context.Context, item operation.Item) error {
	// Get Backup object. It also covers NotFound case
	object, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.client.PlatformV1alpha1().ArangoPlatformServices(item.Namespace).Get, item.Name, meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if object.GetDeletionTimestamp() != nil {
		// We are deleting the object
		finalizer, err := h.finalizer(ctx, object)
		if err != nil {
			return err
		}

		if finalizer != "" {
			if changed, err := patcher.EnsureFinalizersGone(ctx, h.client.PlatformV1alpha1().ArangoPlatformServices(item.Namespace), object, finalizer); err != nil {
				return err
			} else if changed {
				return operator.Reconcile("Finalizers updated")
			}
		}

		return operator.Reconcile("Finalizers pending removal")
	}

	if changed, err := patcher.EnsureFinalizersPresent(ctx, h.client.PlatformV1alpha1().ArangoPlatformServices(item.Namespace), object, platformApi.FinalizerArangoPlatformServiceRelease); err != nil {
		return err
	} else if changed {
		return operator.Reconcile("Finalizers updated")
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

	if _, err := operator.WithArangoPlatformServiceUpdateStatusInterfaceRetry(ctx, h.client.PlatformV1alpha1().ArangoPlatformServices(object.GetNamespace()), object, *status, meta.UpdateOptions{}); err != nil {
		return err
	}

	return reconcileErr
}

func (h *handler) finalizer(ctx context.Context, extension *platformApi.ArangoPlatformService) (string, error) {
	for _, finalizer := range extension.GetFinalizers() {
		switch finalizer {
		case platformApi.FinalizerArangoPlatformServiceRelease:
			// Remove Release
			if _, err := h.helm.Uninstall(ctx, extension.GetName(), func(in *action.Uninstall) {
				in.IgnoreNotFound = true
				in.Wait = true
				in.Timeout = 20 * time.Minute
			}); err != nil {
				return "", err
			}

			return platformApi.FinalizerArangoPlatformServiceRelease, nil
		}
	}

	return "", nil
}

func (h *handler) Timeout() time.Duration {
	return 30 * time.Minute // Timeout of the Helm Install Command
}

func (h *handler) handle(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformService, status *platformApi.ArangoPlatformServiceStatus) (bool, error) {
	return operator.HandleP3WithCondition(ctx, &status.Conditions, platformApi.ReadyCondition, item, extension, status, h.HandleSpecValidity, h.HandleDeployment)
}

func (h *handler) HandleSpecValidity(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformService, status *platformApi.ArangoPlatformServiceStatus) (bool, error) {
	if err := extension.Spec.Validate(); err != nil {
		// We have received an error in the spec!

		logger.Err(err).Warn("Invalid Spec on %s", item.String())

		if status.Conditions.Update(platformApi.SpecValidCondition, false, "Spec is invalid", "Spec is invalid") {
			return true, operator.Stop("Invalid spec")
		}
		return false, operator.Stop("Invalid spec")
	}

	if status.Conditions.Update(platformApi.SpecValidCondition, true, "Spec is valid", "Spec is valid") {
		logger.WrapObj(item).Debug("Spec is valid")
		return true, nil
	}

	return false, nil
}

func (h *handler) HandleDeployment(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformService, status *platformApi.ArangoPlatformServiceStatus) (bool, error) {
	logger := logger.WrapObj(item).Str("deployment", extension.Spec.Deployment.GetName())

	if status.Deployment == nil {
		depl, err := h.client.DatabaseV1().ArangoDeployments(extension.GetNamespace()).Get(ctx, extension.Spec.Deployment.GetName(), meta.GetOptions{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return false, err
			}

			if status.Conditions.Update(platformApi.DeploymentFoundCondition, false, "Deployment not found", "Deployment not found") {
				logger.Warn("Deployment Not Found")
				return true, operator.Reconcile("Conditions updated")
			}

			return false, operator.Stop("Missing deployment")
		}

		status.Deployment = util.NewType(sharedApi.NewObject(depl))

		logger.Info("Deployment Accepted")

		return true, operator.Reconcile("Deployment Accepted")
	}

	depl, err := h.client.DatabaseV1().ArangoDeployments(extension.GetNamespace()).Get(ctx, extension.Status.Deployment.GetName(), meta.GetOptions{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return false, err
		}

		if status.Conditions.Update(platformApi.DeploymentFoundCondition, false, "Deployment not found", "Deployment not found") {
			logger.Warn("Deployment Not Found")
			return true, nil
		}

		return false, operator.Stop("Missing deployment, recreate object")
	}

	if !extension.Status.Deployment.Equals(depl) {
		if status.Conditions.Update(platformApi.DeploymentFoundCondition, false, "Deployment changed", "Deployment changed") {
			logger.Warn("Deployment Changed")
			return true, operator.Reconcile("Conditions updated")
		}

		return false, operator.Stop("Invalid deployment, recreate object")
	}

	if status.Conditions.Update(platformApi.DeploymentFoundCondition, true, "Deployment found", "Deployment found") {
		logger.Debug("Deployment Found")
		return true, nil
	}

	return operator.HandleP4(ctx, item, extension, status, depl, h.HandleChart)
}

func (h *handler) HandleChart(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformService, status *platformApi.ArangoPlatformServiceStatus, depl *api.ArangoDeployment) (bool, error) {
	logger := logger.WrapObj(item).Str("chart", extension.Spec.Chart.GetName())

	if status.Chart == nil {
		// Find the chart
		chart, err := h.client.PlatformV1alpha1().ArangoPlatformCharts(extension.GetNamespace()).Get(ctx, extension.Spec.Chart.GetName(), meta.GetOptions{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return false, err
			}

			if status.Conditions.Update(platformApi.ChartFoundCondition, false, "Chart not found", "Chart not found") {
				logger.Warn("Chart Not Found")
				return true, operator.Reconcile("Condition Changed")
			}
		} else {
			status.Chart = util.NewType(sharedApi.NewObject(chart))
			logger.Info("Chart Accepted")
			return true, operator.Reconcile("Chart Accepted")
		}

		return false, operator.Stop("Chart Not Accepted")
	}

	chart, err := h.client.PlatformV1alpha1().ArangoPlatformCharts(extension.GetNamespace()).Get(ctx, extension.Spec.Chart.GetName(), meta.GetOptions{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return false, err
		}

		status.Chart = nil

		if status.Conditions.Update(platformApi.ChartFoundCondition, false, "Chart not found", "Chart not found") {
			logger.Warn("Chart Gone")
			return true, operator.Reconcile("Condition Changed")
		}

		return true, operator.Reconcile("Chart Gone")
	}

	if !status.Chart.Equals(chart) {
		if !apiErrors.IsNotFound(err) {
			return false, err
		}

		status.Chart = nil

		if status.Conditions.Update(platformApi.ChartFoundCondition, false, "Chart Changed", "Chart Changed") {
			logger.Warn("Chart Changed")
			return true, operator.Reconcile("Condition Changed")
		}

		return true, operator.Reconcile("Chart Changed")
	}

	if status.Conditions.UpdateWithHash(platformApi.ChartFoundCondition, true, "Chart found", "Chart found", chart.Status.Info.GetChecksum()) {
		return true, operator.Reconcile("Condition Changed")
	}

	// Ensure chart is ready
	if !chart.Ready() {
		logger.Warn("Chart is not ready")
		return false, operator.Stop("Chart Not Ready")
	}

	return operator.HandleP5WithCondition(ctx, &status.Conditions, platformApi.ReleaseReadyCondition, item, extension, status, depl, chart, h.HandleChartInfo, h.HandleValues, h.HandleRelease)
}

func (h *handler) HandleChartInfo(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformService, status *platformApi.ArangoPlatformServiceStatus, depl *api.ArangoDeployment, chart *platformApi.ArangoPlatformChart) (bool, error) {
	if chart.Status.Info == nil {
		return false, operator.Stop("Chart Not Ready")
	}

	if status.ChartInfo == nil {
		status.ChartInfo = chart.Status.Info.DeepCopy()

		logger.WrapObj(item).Str("checksum", status.ChartInfo.Checksum).Info("Chart Accepted")

		return true, operator.Reconcile("Chart Changed")
	}

	if status.ChartInfo.Checksum != chart.Status.Info.Checksum {
		status.ChartInfo = nil

		logger.WrapObj(item).Debug("Chart Changed")

		return true, operator.Reconcile("Chart Changed")
	}

	return false, nil
}

func (h *handler) HandleValues(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformService, status *platformApi.ArangoPlatformServiceStatus, depl *api.ArangoDeployment, chart *platformApi.ArangoPlatformChart) (bool, error) {
	vs, err := platform.Service{
		Platform: platform.ServicePlatform{
			Deployment: platform.ServicePlatformDeployment{
				Name: depl.GetName(),
			},
		},
	}.Values()
	if err != nil {
		return false, err
	}

	nw, err := helm.NewMergeRawValues(helm.MergeMaps, vs, helm.Values(extension.Status.ChartInfo.Overrides), helm.Values(extension.Spec.Values))
	if err != nil {
		return false, err
	}

	if !status.Values.Equals(sharedApi.Any(nw)) {
		status.Values = sharedApi.Any(nw)

		logger.WrapObj(item).Str("checksum", status.Values.SHA256()).Info("Values Changed")
		return true, operator.Reconcile("Values Changed")
	}

	return false, nil
}

func (h *handler) HandleRelease(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformService, status *platformApi.ArangoPlatformServiceStatus, depl *api.ArangoDeployment, chart *platformApi.ArangoPlatformChart) (bool, error) {
	expectedChecksum := util.SHA256FromStringArray(status.ChartInfo.Checksum, status.Values.SHA256())

	release, err := h.helm.Status(ctx, extension.Name)
	if err != nil {
		return false, err
	}

	if release == nil {
		// Install
		logger.WrapObj(item).Info("Install Helm Release")
		release, err = h.helm.Install(ctx, helm.Chart(status.ChartInfo.Definition), helm.Values(status.Values), func(in *action.Install) {
			in.ReleaseName = extension.GetName()
			in.Namespace = extension.GetNamespace()

			in.Labels = platform.GetLabels(status.Deployment.GetName(), status.Chart.GetName())

			in.Timeout = 20 * time.Minute
		})
		if err != nil {
			return false, err
		}

		status.Release = extractReleaseStatus(release, expectedChecksum)

		return true, operator.Reconcile("Release Installed")
	} else if !platform.IsPlatformManaged(release) {
		return false, operator.Stop("Release already installed")
	}

	if status.Release == nil || status.Release.Version != release.Version {
		logger.WrapObj(item).Info("Fetch Helm Release Info")

		status.Release = extractReleaseStatus(release, expectedChecksum)

		return true, operator.Reconcile("Release Fetched")
	}

	if status.Release.Hash != expectedChecksum {
		// We need to run an upgrade
		logger.WrapObj(item).Info("Upgrade Helm Release")

		_, err = h.helm.Upgrade(ctx, extension.GetName(), helm.Chart(status.ChartInfo.Definition), helm.Values(status.Values), func(in *action.Upgrade) {
			in.Namespace = extension.GetNamespace()

			in.Labels = platform.GetLabels(status.Deployment.GetName(), status.Chart.GetName())

			in.Timeout = 20 * time.Minute
		})
		if err != nil {
			return false, err
		}

		status.Release = extractReleaseStatus(release, expectedChecksum)
		return true, operator.Reconcile("Release Upgraded")
	}

	return false, nil
}

func extractReleaseStatus(in *helm.Release, hash string) *platformApi.ArangoPlatformServiceStatusRelease {
	if in == nil {
		return nil
	}

	return &platformApi.ArangoPlatformServiceStatusRelease{
		Name:    in.Name,
		Version: in.Version,

		Hash: hash,

		Info: extractReleaseStatusInfo(in.Info),
	}
}

func extractReleaseStatusInfo(in helm.ReleaseInfo) platformApi.ArangoPlatformServiceStatusReleaseInfo {
	var i platformApi.ArangoPlatformServiceStatusReleaseInfo
	if !in.FirstDeployed.IsZero() {
		i.FirstDeployed = util.NewType(meta.NewTime(in.FirstDeployed))
	}
	if !in.LastDeployed.IsZero() {
		i.LastDeployed = util.NewType(meta.NewTime(in.LastDeployed))
	}
	i.Status = in.Status
	return i
}

func (h *handler) CanBeHandled(item operation.Item) bool {
	return item.Group == Group() &&
		item.Version == Version() &&
		item.Kind == Kind()
}
