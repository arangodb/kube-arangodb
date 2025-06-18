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

package v2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"helm.sh/helm/v3/pkg/action"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/suite"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Test_Implementation(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	scheduler, _, _, h := Client(t, ctx, func(c Configuration) Configuration {
		c.Deployment = tests.FakeNamespace
		return c
	})

	values, err := helm.NewValues(map[string]string{
		"A": "B",
	})
	require.NoError(t, err)

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
		_, err := scheduler.Status(context.Background(), &pbSchedulerV2.SchedulerV2StatusRequest{
			Name: "test",
		})
		tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound)
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
			Chart:  suite.GetChart(t, "example", "1.0.0"),
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
		resp, err := h.Install(context.Background(), suite.GetChart(t, "example", "1.0.0"), nil, func(in *action.Install) {
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
			Chart:  suite.GetChart(t, "example", "1.0.0"),
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
		resp, err := h.Install(context.Background(), suite.GetChart(t, "example", "1.0.0"), nil, func(in *action.Install) {
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
		require.Len(t, status.GetRelease().GetValues(), 0)
	})

	t.Run("Upgrade", func(t *testing.T) {
		status, err := scheduler.Upgrade(context.Background(), &pbSchedulerV2.SchedulerV2UpgradeRequest{
			Name:   "test",
			Values: values,
			Chart:  suite.GetChart(t, "example", "1.0.0"),
		})
		require.NoError(t, err)

		require.NotNil(t, status.GetAfter())
		t.Logf("Data: %s", string(status.GetAfter().GetValues()))
		require.Len(t, status.GetAfter().GetValues(), len(values))
	})

	t.Run("Upgrade to 1", func(t *testing.T) {
		status, err := scheduler.Upgrade(context.Background(), &pbSchedulerV2.SchedulerV2UpgradeRequest{
			Name:   "test",
			Values: values,
			Chart:  suite.GetChart(t, "example", "1.0.1"),
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

		require.EqualValues(t, 3, status.GetRelease().GetVersion())
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

func Test_ImplementationV2(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	scheduler, ns, client, _ := Client(t, ctx, func(c Configuration) Configuration {
		c.Deployment = tests.FakeNamespace
		return c
	})

	h := chartHandler(client, ns)

	chart := tests.NewMetaObject[*platformApi.ArangoPlatformChart](t, ns, "secret",
		func(t *testing.T, obj *platformApi.ArangoPlatformChart) {})

	refresh := tests.CreateObjects(t, client.Kubernetes(), client.Arango(), &chart)

	refresh(t)

	t.Run("Check Empty", func(t *testing.T) {
		rs, err := scheduler.List(context.Background(), &pbSchedulerV2.SchedulerV2ListRequest{})
		require.NoError(t, err)
		require.Len(t, rs.Releases, 0)
	})

	t.Run("Install Empty - Missing Chart", func(t *testing.T) {
		_, err := scheduler.InstallV2(context.Background(), &pbSchedulerV2.SchedulerV2InstallV2Request{
			Name:  "example",
			Chart: "missing",
		})
		tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound).Errorf(t, "NotFound: arangoplatformcharts.platform.arangodb.com \"missing\" not found")
	})

	t.Run("Install Empty - Invalid Chart", func(t *testing.T) {
		_, err := scheduler.InstallV2(context.Background(), &pbSchedulerV2.SchedulerV2InstallV2Request{
			Name:  "example",
			Chart: "secret",
		})
		tgrpc.AsGRPCError(t, err).Code(t, codes.Unavailable)
	})

	t.Run("Install", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			tests.Update(t, client.Kubernetes(), client.Arango(), &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
				obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
				obj.Spec.Overrides = nil
			})

			require.NoError(t, tests.Handle(h, tests.NewItem(t, operation.Update, chart)))

			refresh(t)

			_, err := scheduler.InstallV2(context.Background(), &pbSchedulerV2.SchedulerV2InstallV2Request{
				Name:  "example",
				Chart: "secret",
			})
			require.NoError(t, err)

			cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "secret", "example")
			require.NotNil(t, cm)

			require.Equal(t, "PLACEHOLDER", cm.Data)

			_, err = scheduler.Uninstall(ctx, &pbSchedulerV2.SchedulerV2UninstallRequest{
				Name:    "example",
				Options: &pbSchedulerV2.SchedulerV2UninstallRequestOptions{},
			})
			require.NoError(t, err)

			_, err = scheduler.Status(ctx, &pbSchedulerV2.SchedulerV2StatusRequest{
				Name: "example",
			})
			tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound)
		})

		t.Run("From Chart", func(t *testing.T) {
			tests.Update(t, client.Kubernetes(), client.Arango(), &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
				obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
				obj.Spec.Overrides = sharedApi.NewAnyT(t, suite.ConfigMapInput{Data: "chart"})
			})

			require.NoError(t, tests.Handle(h, tests.NewItem(t, operation.Update, chart)))

			refresh(t)

			_, err := scheduler.InstallV2(context.Background(), &pbSchedulerV2.SchedulerV2InstallV2Request{
				Name:  "example",
				Chart: "secret",
			})
			require.NoError(t, err)

			cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "secret", "example")
			require.NotNil(t, cm)

			require.Equal(t, "chart", cm.Data)

			_, err = scheduler.Uninstall(ctx, &pbSchedulerV2.SchedulerV2UninstallRequest{
				Name:    "example",
				Options: &pbSchedulerV2.SchedulerV2UninstallRequestOptions{},
			})
			require.NoError(t, err)

			_, err = scheduler.Status(ctx, &pbSchedulerV2.SchedulerV2StatusRequest{
				Name: "example",
			})
			tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound)
		})

		t.Run("From Service", func(t *testing.T) {
			tests.Update(t, client.Kubernetes(), client.Arango(), &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
				obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
				obj.Spec.Overrides = sharedApi.NewAnyT(t, suite.ConfigMapInput{Data: "chart"})
			})

			require.NoError(t, tests.Handle(h, tests.NewItem(t, operation.Update, chart)))

			refresh(t)

			_, err := scheduler.InstallV2(context.Background(), &pbSchedulerV2.SchedulerV2InstallV2Request{
				Name:  "example",
				Chart: "secret",
				Values: [][]byte{
					sharedApi.NewAnyT(t, suite.ConfigMapInput{Data: "service"}),
				},
			})
			require.NoError(t, err)

			cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "secret", "example")
			require.NotNil(t, cm)

			require.Equal(t, "service", cm.Data)

			_, err = scheduler.Uninstall(ctx, &pbSchedulerV2.SchedulerV2UninstallRequest{
				Name:    "example",
				Options: &pbSchedulerV2.SchedulerV2UninstallRequestOptions{},
			})
			require.NoError(t, err)

			_, err = scheduler.Status(ctx, &pbSchedulerV2.SchedulerV2StatusRequest{
				Name: "example",
			})
			tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound)
		})

		t.Run("From Service over Chart", func(t *testing.T) {
			tests.Update(t, client.Kubernetes(), client.Arango(), &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
				obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
				obj.Spec.Overrides = nil
			})

			require.NoError(t, tests.Handle(h, tests.NewItem(t, operation.Update, chart)))

			refresh(t)

			_, err := scheduler.InstallV2(context.Background(), &pbSchedulerV2.SchedulerV2InstallV2Request{
				Name:  "example",
				Chart: "secret",
				Values: [][]byte{
					sharedApi.NewAnyT(t, suite.ConfigMapInput{Data: "service"}),
				},
			})
			require.NoError(t, err)

			cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "secret", "example")
			require.NotNil(t, cm)

			require.Equal(t, "service", cm.Data)

			_, err = scheduler.Uninstall(ctx, &pbSchedulerV2.SchedulerV2UninstallRequest{
				Name:    "example",
				Options: &pbSchedulerV2.SchedulerV2UninstallRequestOptions{},
			})
			require.NoError(t, err)

			_, err = scheduler.Status(ctx, &pbSchedulerV2.SchedulerV2StatusRequest{
				Name: "example",
			})
			tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound)
		})
	})

	t.Run("Upgrade", func(t *testing.T) {
		t.Run("Install", func(t *testing.T) {
			tests.Update(t, client.Kubernetes(), client.Arango(), &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
				obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
				obj.Spec.Overrides = nil
			})

			require.NoError(t, tests.Handle(h, tests.NewItem(t, operation.Update, chart)))

			refresh(t)

			_, err := scheduler.InstallV2(context.Background(), &pbSchedulerV2.SchedulerV2InstallV2Request{
				Name:  "example",
				Chart: "secret",
			})
			require.NoError(t, err)

			cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "secret", "example")
			require.NotNil(t, cm)

			require.Equal(t, "PLACEHOLDER", cm.Data)
		})

		t.Run("Empty", func(t *testing.T) {
			tests.Update(t, client.Kubernetes(), client.Arango(), &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
				obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
				obj.Spec.Overrides = nil
			})

			require.NoError(t, tests.Handle(h, tests.NewItem(t, operation.Update, chart)))

			refresh(t)

			_, err := scheduler.UpgradeV2(context.Background(), &pbSchedulerV2.SchedulerV2UpgradeV2Request{
				Name:  "example",
				Chart: "secret",
			})
			require.NoError(t, err)

			cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "secret", "example")
			require.NotNil(t, cm)

			require.Equal(t, "PLACEHOLDER", cm.Data)
		})

		t.Run("From Chart", func(t *testing.T) {
			tests.Update(t, client.Kubernetes(), client.Arango(), &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
				obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
				obj.Spec.Overrides = sharedApi.NewAnyT(t, suite.ConfigMapInput{Data: "chart"})
			})

			require.NoError(t, tests.Handle(h, tests.NewItem(t, operation.Update, chart)))

			refresh(t)

			_, err := scheduler.UpgradeV2(context.Background(), &pbSchedulerV2.SchedulerV2UpgradeV2Request{
				Name:  "example",
				Chart: "secret",
			})
			require.NoError(t, err)

			cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "secret", "example")
			require.NotNil(t, cm)

			require.Equal(t, "chart", cm.Data)
		})

		t.Run("From Service", func(t *testing.T) {
			tests.Update(t, client.Kubernetes(), client.Arango(), &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
				obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
				obj.Spec.Overrides = sharedApi.NewAnyT(t, suite.ConfigMapInput{Data: "chart"})
			})

			require.NoError(t, tests.Handle(h, tests.NewItem(t, operation.Update, chart)))

			refresh(t)

			_, err := scheduler.UpgradeV2(context.Background(), &pbSchedulerV2.SchedulerV2UpgradeV2Request{
				Name:  "example",
				Chart: "secret",
				Values: [][]byte{
					sharedApi.NewAnyT(t, suite.ConfigMapInput{Data: "service"}),
				},
			})
			require.NoError(t, err)

			cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "secret", "example")
			require.NotNil(t, cm)

			require.Equal(t, "service", cm.Data)
		})

		t.Run("From Service over Chart", func(t *testing.T) {
			tests.Update(t, client.Kubernetes(), client.Arango(), &chart, func(t *testing.T, obj *platformApi.ArangoPlatformChart) {
				obj.Spec.Definition = suite.GetChart(t, "secret", "1.0.0")
				obj.Spec.Overrides = nil
			})

			require.NoError(t, tests.Handle(h, tests.NewItem(t, operation.Update, chart)))

			refresh(t)

			_, err := scheduler.UpgradeV2(context.Background(), &pbSchedulerV2.SchedulerV2UpgradeV2Request{
				Name:  "example",
				Chart: "secret",
				Values: [][]byte{
					sharedApi.NewAnyT(t, suite.ConfigMapInput{Data: "service"}),
				},
			})
			require.NoError(t, err)

			cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "secret", "example")
			require.NotNil(t, cm)

			require.Equal(t, "service", cm.Data)
		})
	})
}
