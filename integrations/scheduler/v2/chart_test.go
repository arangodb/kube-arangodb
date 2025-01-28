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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Test_Chart_List(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	scheduler, client, _ := InternalClient(t, ctx, func(c Configuration) Configuration {
		c.Namespace = tests.FakeNamespace
		c.Deployment = tests.FakeNamespace
		return c
	})

	t.Run("Create Charts", func(t *testing.T) {
		for i := 0; i < 1024; i++ {
			_, err := client.Arango().PlatformV1alpha1().ArangoPlatformCharts(tests.FakeNamespace).Create(context.Background(), &platformApi.ArangoPlatformChart{
				ObjectMeta: meta.ObjectMeta{
					Name:      fmt.Sprintf("chart-%05d", i),
					Namespace: tests.FakeNamespace,
				},
				Status: platformApi.ArangoPlatformChartStatus{
					Conditions: []api.Condition{
						{
							Type:   platformApi.SpecValidCondition,
							Status: core.ConditionTrue,
						},
					},
					Info: &platformApi.ChartStatusInfo{
						Definition: make([]byte, 128),
						Valid:      true,
						Details: &platformApi.ChartDetails{
							Name:    fmt.Sprintf("chart-%05d", i),
							Version: "1.2.3",
						},
					},
				},
			}, meta.CreateOptions{})
			require.NoError(t, err)
		}
	})

	t.Run("Try to get", func(t *testing.T) {
		_, err := scheduler.GetChart(context.Background(), &pbSchedulerV2.SchedulerV2GetChartRequest{
			Name: "chart-00010",
		})
		require.NoError(t, err)
	})

	t.Run("List by 128", func(t *testing.T) {
		wr, err := scheduler.ListCharts(ctx, &pbSchedulerV2.SchedulerV2ListChartsRequest{
			Items: util.NewType[int64](128),
		})
		require.NoError(t, err)

		var items []string

		require.NoError(t, ugrpc.Recv[*pbSchedulerV2.SchedulerV2ListChartsResponse](wr, func(response *pbSchedulerV2.SchedulerV2ListChartsResponse) error {
			items = append(items, response.GetCharts()...)
			return nil
		}))

		require.Len(t, items, 1024)
	})
}

func Test_Chart_Get(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	scheduler, client, _ := InternalClient(t, ctx, func(c Configuration) Configuration {
		c.Namespace = tests.FakeNamespace
		c.Deployment = tests.FakeNamespace
		return c
	})

	z := client.Arango().PlatformV1alpha1().ArangoPlatformCharts(tests.FakeNamespace)

	t1, err := z.Create(context.Background(), &platformApi.ArangoPlatformChart{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-1",
			Namespace: tests.FakeNamespace,
		},
		Status: platformApi.ArangoPlatformChartStatus{},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	t2, err := z.Create(context.Background(), &platformApi.ArangoPlatformChart{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-2",
			Namespace: tests.FakeNamespace,
		},
		Status: platformApi.ArangoPlatformChartStatus{
			Conditions: []api.Condition{
				{
					Type:   platformApi.SpecValidCondition,
					Status: core.ConditionTrue,
				},
			},
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	t3, err := z.Create(context.Background(), &platformApi.ArangoPlatformChart{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-3",
			Namespace: tests.FakeNamespace,
		},
		Status: platformApi.ArangoPlatformChartStatus{
			Conditions: []api.Condition{
				{
					Type:   platformApi.SpecValidCondition,
					Status: core.ConditionTrue,
				},
			},
			Info: &platformApi.ChartStatusInfo{
				Valid: false,
			},
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	t4, err := z.Create(context.Background(), &platformApi.ArangoPlatformChart{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-4",
			Namespace: tests.FakeNamespace,
		},
		Status: platformApi.ArangoPlatformChartStatus{
			Conditions: []api.Condition{
				{
					Type:   platformApi.SpecValidCondition,
					Status: core.ConditionTrue,
				},
			},
			Info: &platformApi.ChartStatusInfo{
				Valid:   false,
				Message: "Invalid XxX",
			},
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	t5, err := z.Create(context.Background(), &platformApi.ArangoPlatformChart{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-5",
			Namespace: tests.FakeNamespace,
		},
		Status: platformApi.ArangoPlatformChartStatus{
			Conditions: []api.Condition{
				{
					Type:   platformApi.SpecValidCondition,
					Status: core.ConditionTrue,
				},
			},
			Info: &platformApi.ChartStatusInfo{
				Definition: make([]byte, 128),
				Valid:      true,
			},
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	t6, err := z.Create(context.Background(), &platformApi.ArangoPlatformChart{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-6",
			Namespace: tests.FakeNamespace,
		},
		Status: platformApi.ArangoPlatformChartStatus{
			Conditions: []api.Condition{
				{
					Type:   platformApi.SpecValidCondition,
					Status: core.ConditionTrue,
				},
			},
			Info: &platformApi.ChartStatusInfo{
				Definition: make([]byte, 128),
				Valid:      true,
				Details: &platformApi.ChartDetails{
					Name:    "test-6",
					Version: "1.2.3",
				},
			},
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	t.Run("Missing", func(t *testing.T) {
		resp, err := scheduler.GetChart(context.Background(), &pbSchedulerV2.SchedulerV2GetChartRequest{Name: "test-0"})
		tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound).Errorf(t, "NotFound: arangoplatformcharts.platform.arangodb.com \"test-0\" not found")
		require.Nil(t, resp)
	})

	t.Run("Invalid Spec", func(t *testing.T) {
		resp, err := scheduler.GetChart(context.Background(), &pbSchedulerV2.SchedulerV2GetChartRequest{Name: t1.GetName()})
		tgrpc.AsGRPCError(t, err).Code(t, codes.Unavailable).Errorf(t, "Chart Spec is invalid")
		require.Nil(t, resp)
	})

	t.Run("Invalid Info", func(t *testing.T) {
		resp, err := scheduler.GetChart(context.Background(), &pbSchedulerV2.SchedulerV2GetChartRequest{Name: t2.GetName()})
		tgrpc.AsGRPCError(t, err).Code(t, codes.Unavailable).Errorf(t, "Chart Infos are missing")
		require.Nil(t, resp)
	})

	t.Run("Invalid", func(t *testing.T) {
		resp, err := scheduler.GetChart(context.Background(), &pbSchedulerV2.SchedulerV2GetChartRequest{Name: t3.GetName()})
		tgrpc.AsGRPCError(t, err).Code(t, codes.Unavailable).Errorf(t, "Chart is not Valid")
		require.Nil(t, resp)
	})

	t.Run("Invalid with message", func(t *testing.T) {
		resp, err := scheduler.GetChart(context.Background(), &pbSchedulerV2.SchedulerV2GetChartRequest{Name: t4.GetName()})
		tgrpc.AsGRPCError(t, err).Code(t, codes.Unavailable).Errorf(t, "Chart is not Valid: Invalid XxX")
		require.Nil(t, resp)
	})

	t.Run("Valid with missing details", func(t *testing.T) {
		resp, err := scheduler.GetChart(context.Background(), &pbSchedulerV2.SchedulerV2GetChartRequest{Name: t5.GetName()})
		tgrpc.AsGRPCError(t, err).Code(t, codes.Unavailable).Errorf(t, "Chart Details are missing")
		require.Nil(t, resp)
	})

	t.Run("Valid with details", func(t *testing.T) {
		resp, err := scheduler.GetChart(context.Background(), &pbSchedulerV2.SchedulerV2GetChartRequest{Name: t6.GetName()})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Chart, 128)
		require.NotNil(t, resp.Info)
		require.EqualValues(t, "test-6", resp.Info.Name)
		require.EqualValues(t, "1.2.3", resp.Info.Version)
	})
}
