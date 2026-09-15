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
	"testing"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Test_discoverySelector(t *testing.T) {
	sel, err := labels.Parse(discoverySelector("depl-a"))
	require.NoError(t, err)

	// Matches a platform service release owned by the deployment.
	require.True(t, sel.Matches(labels.Set{
		utilConstants.HelmLabelArangoDBManaged:    "true",
		utilConstants.HelmLabelArangoDBType:       utilConstants.HelmTypeService.String(),
		utilConstants.LabelArangoDBDeploymentName: "depl-a",
	}))

	// Rejects a non-managed release.
	require.False(t, sel.Matches(labels.Set{
		utilConstants.HelmLabelArangoDBType:       utilConstants.HelmTypeService.String(),
		utilConstants.LabelArangoDBDeploymentName: "depl-a",
	}))

	// Rejects a platform-typed (non-service) release.
	require.False(t, sel.Matches(labels.Set{
		utilConstants.HelmLabelArangoDBManaged:    "true",
		utilConstants.HelmLabelArangoDBType:       utilConstants.HelmTypePlatform.String(),
		utilConstants.LabelArangoDBDeploymentName: "depl-a",
	}))

	// Rejects a release owned by a different deployment.
	require.False(t, sel.Matches(labels.Set{
		utilConstants.HelmLabelArangoDBManaged:    "true",
		utilConstants.HelmLabelArangoDBType:       utilConstants.HelmTypeService.String(),
		utilConstants.LabelArangoDBDeploymentName: "depl-b",
	}))
}

func Test_ensureDiscoveredWorkflow(t *testing.T) {
	const ns = "test"

	depl := &api.ArangoDeployment{ObjectMeta: meta.ObjectMeta{Name: "depl-a", Namespace: ns}}

	newHandler := func() *handler {
		return &handler{client: kclient.NewFakeClient().Arango()}
	}

	release := func(name, chartLabel string) *helm.Release {
		r := &helm.Release{Name: name, Labels: map[string]string{}}
		if chartLabel != "" {
			r.Labels[utilConstants.HelmLabelArangoDBChart] = chartLabel
		}
		return r
	}

	t.Run("No chart label creates a chart-less workflow", func(t *testing.T) {
		h := newHandler()

		require.NoError(t, h.ensureDiscoveredWorkflow(context.Background(), depl, release("app", "")))

		wf, err := h.client.PlatformV1beta1().ArangoPlatformWorkflows(ns).Get(context.Background(), "app", meta.GetOptions{})
		require.NoError(t, err)
		require.Nil(t, wf.Spec.Chart)
		require.NotNil(t, wf.Spec.Deployment)
		require.Equal(t, "depl-a", wf.Spec.Deployment.GetName())
		require.Equal(t, "depl-a", wf.GetLabels()[utilConstants.LabelArangoDBDeploymentName])
	})

	t.Run("Chart label without a chart CR creates a chart-less workflow", func(t *testing.T) {
		h := newHandler()

		require.NoError(t, h.ensureDiscoveredWorkflow(context.Background(), depl, release("app", "gral")))

		wf, err := h.client.PlatformV1beta1().ArangoPlatformWorkflows(ns).Get(context.Background(), "app", meta.GetOptions{})
		require.NoError(t, err)
		require.Nil(t, wf.Spec.Chart)
	})

	t.Run("Chart label with an existing chart CR adopts the release", func(t *testing.T) {
		h := newHandler()

		_, err := h.client.PlatformV1beta1().ArangoPlatformCharts(ns).Create(context.Background(), &platformApi.ArangoPlatformChart{
			ObjectMeta: meta.ObjectMeta{Name: "gral", Namespace: ns},
		}, meta.CreateOptions{})
		require.NoError(t, err)

		require.NoError(t, h.ensureDiscoveredWorkflow(context.Background(), depl, release("app", "gral")))

		wf, err := h.client.PlatformV1beta1().ArangoPlatformWorkflows(ns).Get(context.Background(), "app", meta.GetOptions{})
		require.NoError(t, err)
		require.NotNil(t, wf.Spec.Chart)
		require.Equal(t, "gral", wf.Spec.Chart.GetName())
	})

	t.Run("Existing workflow is left untouched", func(t *testing.T) {
		h := newHandler()

		_, err := h.client.PlatformV1beta1().ArangoPlatformWorkflows(ns).Create(context.Background(), &platformApi.ArangoPlatformWorkflow{
			ObjectMeta: meta.ObjectMeta{Name: "app", Namespace: ns},
		}, meta.CreateOptions{})
		require.NoError(t, err)

		// Even with a resolvable chart label, an existing workflow must not be modified.
		_, err = h.client.PlatformV1beta1().ArangoPlatformCharts(ns).Create(context.Background(), &platformApi.ArangoPlatformChart{
			ObjectMeta: meta.ObjectMeta{Name: "gral", Namespace: ns},
		}, meta.CreateOptions{})
		require.NoError(t, err)

		require.NoError(t, h.ensureDiscoveredWorkflow(context.Background(), depl, release("app", "gral")))

		wf, err := h.client.PlatformV1beta1().ArangoPlatformWorkflows(ns).Get(context.Background(), "app", meta.GetOptions{})
		require.NoError(t, err)
		require.Nil(t, wf.Spec.Chart)
	})
}
