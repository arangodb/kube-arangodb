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

package v2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/action"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func cleanup(t *testing.T, c helm.Client) func() {
	t.Run("Cleanup Pre", func(t *testing.T) {
		items, err := c.List(context.Background(), func(in *action.List) {
			in.All = true
			in.StateMask = action.ListDeployed | action.ListUninstalled | action.ListUninstalling | action.ListPendingInstall | action.ListPendingUpgrade | action.ListPendingRollback | action.ListSuperseded | action.ListFailed | action.ListUnknown
		})
		require.NoError(t, err)

		for _, item := range items {
			t.Run(item.Name, func(t *testing.T) {
				_, err := c.Uninstall(context.Background(), item.Name)
				require.NoError(t, err)
			})
		}
	})

	return func() {
		t.Run("Cleanup Post", func(t *testing.T) {
			items, err := c.List(context.Background(), func(in *action.List) {
				in.All = true
				in.StateMask = action.ListDeployed | action.ListUninstalled | action.ListUninstalling | action.ListPendingInstall | action.ListPendingUpgrade | action.ListPendingRollback | action.ListSuperseded | action.ListFailed | action.ListUnknown
			})
			require.NoError(t, err)

			for _, item := range items {
				t.Run(item.Name, func(t *testing.T) {
					_, err := c.Uninstall(context.Background(), item.Name)
					require.NoError(t, err)
				})
			}
		})
	}
}

func Test_Implementation(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	scheduler, h := ExternalClient(t, ctx, func(c Configuration) Configuration {
		c.Namespace = tests.FakeNamespace
		c.Deployment = tests.FakeNamespace
		return c
	})

	values, err := helm.NewValues(map[string]string{
		"A": "B",
	})
	require.NoError(t, err)

	defer cleanup(t, h)()

	t.Run("Alive", func(t *testing.T) {
		_, err := scheduler.Alive(context.Background(), &pbSharedV1.Empty{})
		require.NoError(t, err)
	})

	t.Run("Check API Resources", func(t *testing.T) {
		o, err := scheduler.DiscoverAPIResources(context.Background(), &pbSchedulerV2.SchedulerV2DiscoverAPIResourcesRequest{
			Version: "v1",
		})
		require.NoError(t, err)
		for _, z := range o.Resources {
			t.Logf("Kind: %s", z.GetKind())
		}
	})

	t.Run("Check API Resource", func(t *testing.T) {

		oz, err := scheduler.DiscoverAPIResource(context.Background(), &pbSchedulerV2.SchedulerV2DiscoverAPIResourceRequest{
			Version: "v1",
			Kind:    "ConfigMap",
		})
		require.NoError(t, err)
		require.NotNil(t, oz.GetResource())
	})

	t.Run("Check API Resource - Missing", func(t *testing.T) {

		oz, err := scheduler.DiscoverAPIResource(context.Background(), &pbSchedulerV2.SchedulerV2DiscoverAPIResourceRequest{
			Version: "v1",
			Kind:    "ConfigMap2",
		})
		require.NoError(t, err)
		require.Nil(t, oz.GetResource())
	})

	t.Run("Status on Missing", func(t *testing.T) {
		status, err := scheduler.Status(context.Background(), &pbSchedulerV2.SchedulerV2StatusRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.Nil(t, status.GetRelease())
	})

	t.Run("List Empty", func(t *testing.T) {
		status, err := scheduler.List(context.Background(), &pbSchedulerV2.SchedulerV2ListRequest{})
		require.NoError(t, err)

		require.Len(t, status.GetReleases(), 0)
	})

	t.Run("Install", func(t *testing.T) {
		status, err := scheduler.Install(context.Background(), &pbSchedulerV2.SchedulerV2InstallRequest{
			Name:   "test",
			Values: nil,
			Chart:  example_1_0_0,
		})
		require.NoError(t, err)

		require.NotNil(t, status.GetRelease())
	})

	t.Run("List After", func(t *testing.T) {
		status, err := scheduler.List(context.Background(), &pbSchedulerV2.SchedulerV2ListRequest{})
		require.NoError(t, err)

		require.Len(t, status.GetReleases(), 1)
	})

	t.Run("Install Outside", func(t *testing.T) {
		resp, err := h.Install(context.Background(), example_1_0_0, nil, func(in *action.Install) {
			in.ReleaseName = "test-outside"
		})
		require.NoError(t, err)

		require.NotNil(t, resp)
	})

	t.Run("List After - Still should not see first one", func(t *testing.T) {
		status, err := scheduler.List(context.Background(), &pbSchedulerV2.SchedulerV2ListRequest{})
		require.NoError(t, err)

		require.Len(t, status.GetReleases(), 1)
	})

	t.Run("Install Second", func(t *testing.T) {
		status, err := scheduler.Install(context.Background(), &pbSchedulerV2.SchedulerV2InstallRequest{
			Name:   "test-x",
			Values: nil,
			Chart:  example_1_0_0,
			Options: &pbSchedulerV2.SchedulerV2InstallRequestOptions{
				Labels: map[string]string{
					"X": "X",
				},
			},
		})
		require.NoError(t, err)

		require.NotNil(t, status.GetRelease())
	})

	t.Run("Install Second Outside", func(t *testing.T) {
		resp, err := h.Install(context.Background(), example_1_0_0, nil, func(in *action.Install) {
			in.ReleaseName = "test-outside-x"
			in.Labels = map[string]string{
				"X": "X",
			}
		})
		require.NoError(t, err)

		require.NotNil(t, resp)
	})

	t.Run("List After - Should see 2 services", func(t *testing.T) {
		status, err := scheduler.List(context.Background(), &pbSchedulerV2.SchedulerV2ListRequest{})
		require.NoError(t, err)

		require.Len(t, status.GetReleases(), 2)
	})

	t.Run("List After - Filter one", func(t *testing.T) {
		status, err := scheduler.List(context.Background(), &pbSchedulerV2.SchedulerV2ListRequest{
			Options: &pbSchedulerV2.SchedulerV2ListRequestOptions{
				Selectors: map[string]string{
					"X": "X",
				},
			},
		})
		require.NoError(t, err)

		require.Len(t, status.GetReleases(), 1)
	})

	t.Run("Check - Version 1", func(t *testing.T) {
		status, err := scheduler.Status(context.Background(), &pbSchedulerV2.SchedulerV2StatusRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.NotNil(t, status.GetRelease())

		require.EqualValues(t, 1, status.GetRelease().GetVersion())
		t.Logf("Data: %s", string(status.GetRelease().GetValues()))
		require.Len(t, status.GetRelease().GetValues(), 4)
	})

	t.Run("Upgrade", func(t *testing.T) {
		status, err := scheduler.Upgrade(context.Background(), &pbSchedulerV2.SchedulerV2UpgradeRequest{
			Name:   "test",
			Values: values,
			Chart:  example_1_0_0,
		})
		require.NoError(t, err)

		require.NotNil(t, status.GetAfter())
		t.Logf("Data: %s", string(status.GetAfter().GetValues()))
		require.Len(t, status.GetAfter().GetValues(), len(values))
	})

	t.Run("Check - Version 2", func(t *testing.T) {
		status, err := scheduler.Status(context.Background(), &pbSchedulerV2.SchedulerV2StatusRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.NotNil(t, status.GetRelease())

		require.EqualValues(t, 2, status.GetRelease().GetVersion())
		t.Logf("Data: %s", string(status.GetRelease().GetValues()))
		require.Len(t, status.GetRelease().GetValues(), len(values))
	})

	t.Run("Test", func(t *testing.T) {
		status, err := scheduler.Test(context.Background(), &pbSchedulerV2.SchedulerV2TestRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.NotNil(t, status.GetRelease())
	})

	t.Run("Uninstall", func(t *testing.T) {
		status, err := scheduler.Uninstall(context.Background(), &pbSchedulerV2.SchedulerV2UninstallRequest{
			Name: "test",
		})
		require.NoError(t, err)

		require.NotNil(t, status.GetRelease())
		t.Logf("Response: %s", status.GetInfo())
	})
}
