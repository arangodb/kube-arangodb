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

package v2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Test_Workflow(t *testing.T) {
	*features.SchedulerV2Workflow().EnabledPointer() = true
	t.Cleanup(func() { features.SchedulerV2Workflow().Reset() })

	ctx, c := context.WithCancel(context.Background())
	defer c()

	scheduler, ns, client, _ := MockClient(t, ctx, func(c Configuration) Configuration {
		c.Deployment = "test-deployment"
		return c
	})

	values, err := helm.NewValues(map[string]string{"A": "B"})
	require.NoError(t, err)

	workflows := client.Arango().PlatformV1beta1().ArangoPlatformWorkflows(ns)

	t.Run("InstallV2 creates the workflow instead of calling Helm", func(t *testing.T) {
		_, err := scheduler.InstallV2(ctx, &pbSchedulerV2.SchedulerV2InstallV2Request{
			Name:   "app",
			Chart:  "chart",
			Values: [][]byte{values},
		})
		require.NoError(t, err)

		wf, err := workflows.Get(ctx, "app", meta.GetOptions{})
		require.NoError(t, err)

		require.NotNil(t, wf.Spec.Deployment)
		require.Equal(t, "test-deployment", wf.Spec.Deployment.Name)
		require.NotNil(t, wf.Spec.Chart)
		require.Equal(t, "chart", wf.Spec.Chart.Name)
		require.NotEmpty(t, wf.Spec.Values)
		require.Equal(t, "test-deployment", wf.GetLabels()[LabelArangoDBDeploymentName])
	})

	t.Run("UpgradeV2 updates the workflow spec", func(t *testing.T) {
		_, err := scheduler.UpgradeV2(ctx, &pbSchedulerV2.SchedulerV2UpgradeV2Request{
			Name:  "app",
			Chart: "chart-v2",
			Options: &pbSchedulerV2.SchedulerV2UpgradeV2RequestOptions{
				MaxHistory: util.NewType[int32](3),
			},
		})
		require.NoError(t, err)

		wf, err := workflows.Get(ctx, "app", meta.GetOptions{})
		require.NoError(t, err)

		require.Equal(t, "chart-v2", wf.Spec.Chart.Name)
		require.NotNil(t, wf.Spec.Upgrade)
		require.NotNil(t, wf.Spec.Upgrade.MaxHistory)
		require.Equal(t, 3, *wf.Spec.Upgrade.MaxHistory)
	})

	t.Run("Uninstall removes the workflow", func(t *testing.T) {
		_, err := scheduler.Uninstall(ctx, &pbSchedulerV2.SchedulerV2UninstallRequest{
			Name: "app",
		})
		require.NoError(t, err)

		_, err = workflows.Get(ctx, "app", meta.GetOptions{})
		require.True(t, apiErrors.IsNotFound(err))
	})

	t.Run("Uninstall of a missing workflow is NotFound", func(t *testing.T) {
		_, err := scheduler.Uninstall(ctx, &pbSchedulerV2.SchedulerV2UninstallRequest{
			Name: "missing",
		})
		tgrpc.AsGRPCError(t, err).Code(t, codes.NotFound)
	})
}
