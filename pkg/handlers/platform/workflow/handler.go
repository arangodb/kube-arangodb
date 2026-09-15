//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package workflow

import (
	"context"
	"time"

	"helm.sh/helm/v3/pkg/action"
	helmRelease "helm.sh/helm/v3/pkg/release"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1/types"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/platform/labels"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
)

var logger = logging.Global().RegisterAndGetLogger("platform-workflow-operator", logging.Info)

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
	object, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.client.PlatformV1beta1().ArangoPlatformWorkflows(item.Namespace).Get, item.Name, meta.GetOptions{})
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
			if changed, err := patcher.EnsureFinalizersGone(ctx, h.client.PlatformV1beta1().ArangoPlatformWorkflows(item.Namespace), object, finalizer); err != nil {
				return err
			} else if changed {
				return operator.Reconcile("Finalizers updated")
			}
		}

		return operator.Reconcile("Finalizers pending removal")
	}

	if changed, err := patcher.EnsureFinalizersPresent(ctx, h.client.PlatformV1beta1().ArangoPlatformWorkflows(item.Namespace), object, platformApi.FinalizerArangoPlatformWorkflowRelease); err != nil {
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

	if _, err := operator.WithArangoPlatformWorkflowUpdateStatusInterfaceRetry(ctx, h.client.PlatformV1beta1().ArangoPlatformWorkflows(object.GetNamespace()), object, *status, meta.UpdateOptions{}); err != nil {
		return err
	}

	return reconcileErr
}

func (h *handler) finalizer(ctx context.Context, extension *platformApi.ArangoPlatformWorkflow) (string, error) {
	for _, finalizer := range extension.GetFinalizers() {
		switch finalizer {
		case platformApi.FinalizerArangoPlatformWorkflowRelease:
			// A discovered (chart-less) workflow only reflects an existing release; it does not own it, so
			// it must NOT uninstall the release on delete.
			if extension.Spec.Chart != nil {
				if _, err := h.helm.Uninstall(ctx, extension.GetName(), func(in *action.Uninstall) {
					in.IgnoreNotFound = true
					in.Wait = true
					in.Timeout = 20 * time.Minute
				}); err != nil {
					return "", err
				}
			}

			return platformApi.FinalizerArangoPlatformWorkflowRelease, nil
		}
	}

	return "", nil
}

func (h *handler) Timeout() time.Duration {
	return 30 * time.Minute // Timeout of the Helm Install Command
}

func (h *handler) handle(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformWorkflow, status *platformApi.ArangoPlatformWorkflowStatus) (bool, error) {
	return operator.HandleP3WithCondition(ctx, &status.Conditions, platformApi.ReadyCondition, item, extension, status, h.HandleSpecValidity, h.HandleDeployment)
}

func (h *handler) HandleSpecValidity(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformWorkflow, status *platformApi.ArangoPlatformWorkflowStatus) (bool, error) {
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

func (h *handler) HandleDeployment(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformWorkflow, status *platformApi.ArangoPlatformWorkflowStatus) (bool, error) {
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

	// No Chart provided: adopt and reflect an existing Helm release in the status instead of installing
	// one. This is intrinsic reconcile behavior for a chart-less workflow; whether such workflows get
	// created (the discovery loop that scans releases by label) is gated separately by its own feature.
	if extension.Spec.Chart == nil {
		return operator.HandleP4(ctx, item, extension, status, depl, h.HandleDiscovery)
	}

	return operator.HandleP4(ctx, item, extension, status, depl, h.HandleChart)
}

// HandleDiscovery reflects an existing Helm release (installed out-of-band, e.g. by the SchedulerV2
// sidecar) into the workflow status without owning its lifecycle. Used when the workflow has no Chart.
func (h *handler) HandleDiscovery(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformWorkflow, status *platformApi.ArangoPlatformWorkflowStatus, depl *api.ArangoDeployment) (bool, error) {
	release, err := h.helm.Status(ctx, extension.GetName())
	if err != nil {
		return false, err
	}

	if release == nil {
		status.Release = nil
		changed := status.Conditions.Update(platformApi.DiscoveredCondition, false, "Release not found", "Release not found")
		changed = status.Conditions.Update(platformApi.ReleaseReadyCondition, false, "Release not found", "Release not found") || changed
		if changed {
			logger.WrapObj(item).Warn("Release to discover not found")
			return true, operator.Reconcile("Condition Changed")
		}
		return false, operator.Stop("Release not discovered")
	}

	// Only adopt releases we actually manage for this deployment: the release must be tagged as
	// platform-managed and carry the same deployment name. A same-named release that we do not manage (or
	// that belongs to another deployment) must not be reflected as if it were ours.
	if !labels.IsPlatformManaged(release) || labels.DeploymentName(release) != status.Deployment.GetName() {
		status.Release = nil
		changed := status.Conditions.Update(platformApi.DiscoveredCondition, false, "Release not managed", "Release not managed")
		changed = status.Conditions.Update(platformApi.ReleaseReadyCondition, false, "Release not managed", "Release not managed") || changed
		if changed {
			logger.WrapObj(item).Str("release", release.Name).Warn("Release to discover is not managed for this deployment")
			return true, operator.Reconcile("Condition Changed")
		}
		return false, operator.Stop("Release not managed")
	}

	changed := false

	if s := extractReleaseStatus(release, ""); !status.Release.Compare(s) {
		status.Release = s
		changed = true
	}

	// Link the source ArangoPlatformChart when the release carries the chart label and the chart CR still
	// exists, so a discovered workflow references the same chart object as a chart-owning one.
	if chartName := labels.Chart(release); chartName != "" {
		chart, err := h.client.PlatformV1beta1().ArangoPlatformCharts(extension.GetNamespace()).Get(ctx, chartName, meta.GetOptions{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return false, err
			}

			// The chart CR is gone; discovery still works from the release metadata below.
			if status.Chart != nil {
				status.Chart = nil
				changed = true
			}
		} else if status.Chart == nil || !status.Chart.Equals(chart) {
			status.Chart = util.NewType(sharedApi.NewObject(chart))
			changed = true
		}
	} else if status.Chart != nil {
		status.Chart = nil
		changed = true
	}

	// Reflect the chart the release was installed from (name + version), discovered from the Helm release
	// metadata. This is the actual Helm chart, independent of the linked ArangoPlatformChart above.
	if c := release.GetChart().GetMetadata(); c != nil {
		if status.ChartInfo == nil || status.ChartInfo.Details == nil ||
			status.ChartInfo.Details.Name != c.GetName() || status.ChartInfo.Details.Version != c.GetVersion() {
			status.ChartInfo = &platformApi.ChartStatusInfo{
				Valid: true,
				Details: &platformApi.ChartDetails{
					Name:    c.GetName(),
					Version: c.GetVersion(),
				},
			}
			changed = true
		}
	}

	if status.Conditions.Update(platformApi.DiscoveredCondition, true, "Release discovered", "Release discovered") {
		changed = true
	}

	ready := release.Info.Status == helmRelease.StatusDeployed
	if status.Conditions.Update(platformApi.ReleaseReadyCondition, ready, "Release "+string(release.Info.Status), "Release "+string(release.Info.Status)) {
		changed = true
	}

	if changed {
		logger.WrapObj(item).Str("release", release.Name).Info("Release Discovered")
		return true, operator.Reconcile("Discovery updated")
	}

	return false, nil
}

func (h *handler) HandleChart(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformWorkflow, status *platformApi.ArangoPlatformWorkflowStatus, depl *api.ArangoDeployment) (bool, error) {
	logger := logger.WrapObj(item).Str("chart", extension.Spec.Chart.GetName())

	if status.Chart == nil {
		// Find the chart
		chart, err := h.client.PlatformV1beta1().ArangoPlatformCharts(extension.GetNamespace()).Get(ctx, extension.Spec.Chart.GetName(), meta.GetOptions{})
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

	chart, err := h.client.PlatformV1beta1().ArangoPlatformCharts(extension.GetNamespace()).Get(ctx, extension.Spec.Chart.GetName(), meta.GetOptions{})
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

func (h *handler) HandleChartInfo(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformWorkflow, status *platformApi.ArangoPlatformWorkflowStatus, depl *api.ArangoDeployment, chart *platformApi.ArangoPlatformChart) (bool, error) {
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

func (h *handler) HandleValues(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformWorkflow, status *platformApi.ArangoPlatformWorkflowStatus, depl *api.ArangoDeployment, chart *platformApi.ArangoPlatformChart) (bool, error) {
	vs, err := types.Service{
		Platform: types.ServicePlatform{
			Deployment: types.ServicePlatformDeployment{
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

func (h *handler) HandleRelease(ctx context.Context, item operation.Item, extension *platformApi.ArangoPlatformWorkflow, status *platformApi.ArangoPlatformWorkflowStatus, depl *api.ArangoDeployment, chart *platformApi.ArangoPlatformChart) (bool, error) {
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

			in.Labels = labels.GetLabels(status.Deployment.GetName(), status.Chart.GetName(), labels.WithDeploymentName(status.Deployment.GetName()), labels.WithType(utilConstants.HelmTypeService))
		}, withInstallActionOverrides(extension.Spec.Install))
		if err != nil {
			h.eventRecorder.Warning(extension, "Release Install Failed", "Release Install failed: %s", err.Error())
			return false, err
		}

		status.Release = extractReleaseStatus(release, expectedChecksum)

		h.eventRecorder.Normal(extension, "Release Installed", "Release installed with version %d on chart %s (%s)", status.Release.Version, status.ChartInfo.Details.Name, status.ChartInfo.Details.Version)

		return true, operator.Reconcile("Release Installed")
	} else if !labels.IsPlatformManaged(release) {
		return false, operator.Stop("Release already installed")
	}

	// Recover from a release left in a pending state by an interrupted install/upgrade/rollback
	// (e.g. an operator restart mid-operation). Helm refuses any further operation on it
	// ("another operation (install/upgrade/rollback) is in progress") until the pending state is
	// cleared, so the reconcile would otherwise retry the upgrade forever.
	switch release.Info.Status {
	case helmRelease.StatusPendingInstall:
		// The initial install never completed and has no previous revision to roll back to - remove it
		// so it can be re-installed on the next reconcile.
		logger.WrapObj(item).Warn("Release stuck in %s, uninstalling to recover", release.Info.Status)
		if _, err := h.helm.Uninstall(ctx, extension.GetName(), func(in *action.Uninstall) {
			in.IgnoreNotFound = true
		}); err != nil {
			h.eventRecorder.Warning(extension, "Release Recovery Failed", "Failed to recover release stuck in %s: %s", release.Info.Status, err.Error())
			return false, err
		}

		h.eventRecorder.Normal(extension, "Release Recovered", "Uninstalled release stuck in %s", release.Info.Status)
		return true, operator.Reconcile("Recovered pending release")
	case helmRelease.StatusPendingUpgrade, helmRelease.StatusPendingRollback:
		// Roll back to the last deployed revision to clear the pending state; the next reconcile will
		// re-run the upgrade if it is still needed.
		logger.WrapObj(item).Warn("Release stuck in %s, rolling back to recover", release.Info.Status)
		if err := h.helm.Rollback(ctx, extension.GetName()); err != nil {
			h.eventRecorder.Warning(extension, "Release Recovery Failed", "Failed to recover release stuck in %s: %s", release.Info.Status, err.Error())
			return false, err
		}

		h.eventRecorder.Normal(extension, "Release Recovered", "Rolled back release stuck in %s", release.Info.Status)
		return true, operator.Reconcile("Recovered pending release")
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

			in.Labels = labels.GetLabels(status.Deployment.GetName(), status.Chart.GetName(), labels.WithDeploymentName(status.Deployment.GetName()), labels.WithType(utilConstants.HelmTypeService))
		}, withUpgradeActionOverrides(extension.Spec.Upgrade))
		if err != nil {
			h.eventRecorder.Warning(extension, "Release Upgrade Failed", "Release upgrade failed: %s", err.Error())

			return false, err
		}

		status.Release = extractReleaseStatus(release, expectedChecksum)

		h.eventRecorder.Normal(extension, "Release Upgraded", "Release upgraded with version %d on chart %s (%s)", status.Release.Version, status.ChartInfo.Details.Name, status.ChartInfo.Details.Version)

		return true, operator.Reconcile("Release Upgraded")
	}

	if s := extractReleaseStatus(release, expectedChecksum); !s.Compare(status.Release) {
		logger.WrapObj(item).Info("Release Update")
		status.Release = s
		return true, operator.Reconcile("Release Updated")
	}

	switch status.Release.Info.Status {
	case helmRelease.StatusDeployed:
		return false, nil

	case helmRelease.StatusUnknown:
		return false, operator.Stop("Invalid release status: %s", status.Release.Info.Status)

	default:
		// Try to upgrade
		logger.WrapObj(item).Info("Upgrade Helm Release")

		_, err = h.helm.Upgrade(ctx, extension.GetName(), helm.Chart(status.ChartInfo.Definition), helm.Values(status.Values), func(in *action.Upgrade) {
			in.Namespace = extension.GetNamespace()

			in.Labels = labels.GetLabels(status.Deployment.GetName(), status.Chart.GetName(), labels.WithDeploymentName(status.Deployment.GetName()), labels.WithType(utilConstants.HelmTypeService))
		}, withUpgradeActionOverrides(extension.Spec.Upgrade))
		if err != nil {
			h.eventRecorder.Warning(extension, "Release Upgrade Failed", "Release upgrade failed: %s", err.Error())

			return false, err
		}

		status.Release = extractReleaseStatus(release, expectedChecksum)

		h.eventRecorder.Normal(extension, "Release Upgraded", "Release upgraded with version %d on chart %s (%s)", status.Release.Version, status.ChartInfo.Details.Name, status.ChartInfo.Details.Version)

		return true, operator.Reconcile("Release Upgraded")
	}
}

func extractReleaseStatus(in *helm.Release, hash string) *platformApi.ArangoPlatformWorkflowStatusRelease {
	if in == nil {
		return nil
	}

	return &platformApi.ArangoPlatformWorkflowStatusRelease{
		Name:    in.Name,
		Version: in.Version,

		Hash: hash,

		Info: extractReleaseStatusInfo(in.Info),
	}
}

func extractReleaseStatusInfo(in helm.ReleaseInfo) platformApi.ArangoPlatformWorkflowStatusReleaseInfo {
	var i platformApi.ArangoPlatformWorkflowStatusReleaseInfo
	if !in.FirstDeployed.IsZero() {
		i.FirstDeployed = util.NewType(meta.NewTime(in.FirstDeployed))
	}
	if !in.LastDeployed.IsZero() {
		i.LastDeployed = util.NewType(meta.NewTime(in.LastDeployed))
	}
	i.Status = in.Status
	i.Description = in.Description
	return i
}

func (h *handler) CanBeHandled(item operation.Item) bool {
	return item.Group == Group() &&
		utilConstants.Version(Version()).IsCompatible(utilConstants.Version(item.Version)) &&
		item.Kind == Kind()
}
