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
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	platformLabels "github.com/arangodb/kube-arangodb/pkg/platform/labels"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

// discoveryInterval is how often the operator scans for platform service Helm releases that have no
// owning ArangoPlatformWorkflow yet. Mirrors the ArangoBackup discovery loop.
const discoveryInterval = 2 * time.Minute

// platformValuesKey is the top-level values key the operator injects into a managed release (the
// `arangodb_platform` section, see types.Service). It is stripped from discovered values so the workflow
// carries only the user-provided values; the operator re-injects the platform section on reconcile.
const platformValuesKey = "arangodb_platform"

// Start launches the discovery loop as a background goroutine. Registered via operator.RegisterStarter.
func (h *handler) Start(stopCh <-chan struct{}) {
	go h.runDiscovery(stopCh)
}

func (h *handler) runDiscovery(stopCh <-chan struct{}) {
	t := time.NewTicker(discoveryInterval)
	defer t.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-t.C:
			if !features.PlatformWorkflowDiscovery().Enabled() {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), discoveryInterval)
			if err := h.discover(ctx); err != nil {
				logger.Err(err).Warn("ArangoPlatformWorkflow discovery failed")
			}
			cancel()
		}
	}
}

// discover scans every ArangoDeployment in the operator namespace for platform service Helm releases and
// creates a chart-less ArangoPlatformWorkflow for each release that is not yet owned by one. The reconcile
// handler then adopts the release and reflects it in the workflow status.
func (h *handler) discover(ctx context.Context) error {
	ns := h.operator.Namespace()

	deployments, err := h.client.DatabaseV1().ArangoDeployments(ns).List(ctx, meta.ListOptions{})
	if err != nil {
		return err
	}

	for id := range deployments.Items {
		if err := h.discoverDeployment(ctx, &deployments.Items[id]); err != nil {
			logger.Err(err).Str("deployment", deployments.Items[id].GetName()).Warn("Unable to discover workflows")
		}
	}

	return nil
}

func (h *handler) discoverDeployment(ctx context.Context, depl *api.ArangoDeployment) error {
	releases, err := h.helm.List(ctx, func(in *action.List) {
		in.Selector = discoverySelector(depl.GetName())
	})
	if err != nil {
		return err
	}

	for id := range releases {
		if err := h.ensureDiscoveredWorkflow(ctx, depl, &releases[id]); err != nil {
			return err
		}
	}

	return nil
}

// discoverySelector selects the platform service releases owned by the given deployment: those we manage,
// tagged as service type, and carrying the deployment-name label the SchedulerV2 integration sets.
func discoverySelector(deployment string) string {
	s := labels.NewSelector()

	for key, value := range map[string]string{
		utilConstants.HelmLabelArangoDBManaged:    "true",
		utilConstants.HelmLabelArangoDBType:       utilConstants.HelmTypeService.String(),
		utilConstants.LabelArangoDBDeploymentName: deployment,
	} {
		r, err := labels.NewRequirement(key, selection.DoubleEquals, []string{value})
		if err != nil {
			logger.Err(err).Warn("Unable to build discovery selector requirement")
			continue
		}
		s = s.Add(*r)
	}

	return s.String()
}

func (h *handler) ensureDiscoveredWorkflow(ctx context.Context, depl *api.ArangoDeployment, release *helm.Release) error {
	// The workflow name is 1:1 with the release name - the reconcile handler looks the release up by the
	// workflow name.
	name := release.Name

	if _, err := h.client.PlatformV1beta1().ArangoPlatformWorkflows(depl.GetNamespace()).Get(ctx, name, meta.GetOptions{}); err == nil {
		// Already owned by a workflow.
		return nil
	} else if !apiErrors.IsNotFound(err) {
		return err
	}

	spec := platformApi.ArangoPlatformWorkflowSpec{
		Deployment: util.NewType(sharedApi.NewObject(depl)),
	}

	// Discover the release's current values, stripping the operator-injected platform section so the
	// workflow carries only the user-provided values. The operator re-injects the platform section on
	// reconcile, so an adopted release keeps its values and is not reset on the first upgrade.
	if values, ok := discoverReleaseValues(release); ok {
		spec.Values = values
	}

	// If the release records the source chart (chart label) and that ArangoPlatformChart exists, adopt the
	// release as a fully managed workflow by referencing the chart. Otherwise create a chart-less workflow
	// that only reflects the release in its status.
	if chartName := platformLabels.Chart(release); chartName != "" {
		chart, err := h.client.PlatformV1beta1().ArangoPlatformCharts(depl.GetNamespace()).Get(ctx, chartName, meta.GetOptions{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
		} else {
			spec.Chart = util.NewType(sharedApi.NewObject(chart))
		}
	}

	wf := &platformApi.ArangoPlatformWorkflow{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: depl.GetNamespace(),
			Labels: map[string]string{
				utilConstants.LabelArangoDBDeploymentName: depl.GetName(),
			},
		},
		Spec: spec,
	}

	if _, err := h.client.PlatformV1beta1().ArangoPlatformWorkflows(depl.GetNamespace()).Create(ctx, wf, meta.CreateOptions{}); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	logger.Str("workflow", name).Str("deployment", depl.GetName()).Info("Discovered ArangoPlatformWorkflow from Helm release")

	return nil
}

// discoverReleaseValues returns the release's user-provided values with the operator-injected platform
// section removed, or ok=false when the release has no user values to carry over.
func discoverReleaseValues(release *helm.Release) (sharedApi.Any, bool) {
	if len(release.Values) == 0 {
		return nil, false
	}

	m, err := release.Values.Marshal()
	if err != nil {
		logger.Err(err).Warn("Unable to read discovered release values")
		return nil, false
	}

	delete(m, platformValuesKey)

	if len(m) == 0 {
		return nil, false
	}

	vs, err := helm.NewValues(m)
	if err != nil {
		logger.Err(err).Warn("Unable to encode discovered release values")
		return nil, false
	}

	return sharedApi.Any(vs), true
}
