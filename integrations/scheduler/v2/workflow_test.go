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
	authorizationApi "k8s.io/api/authorization/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetesFake "k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Test_Workflow(t *testing.T) {
	*features.SchedulerV2Workflow().EnabledPointer() = true
	t.Cleanup(func() { features.SchedulerV2Workflow().Reset() })

	ctx, c := context.WithCancel(context.Background())
	defer c()

	scheduler, ns, client, _ := MockClient(t, ctx, func(c Configuration) Configuration {
		c.Deployment = "test-deployment"
		// Mirror the sidecar default: MaxHistory is not set via a flag, so it is 0. The CRD requires >= 1,
		// so the workflow spec must not carry it unless an explicit option is provided.
		c.MaxHistory = 0
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
		require.Equal(t, "test-deployment", wf.GetLabels()[utilConstants.LabelArangoDBDeploymentName])
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

	t.Run("UpgradeV2 without MaxHistory defaults to a valid value", func(t *testing.T) {
		// The configured MaxHistory may be unset (0); the CRD requires >= 1, so with no option the sidecar
		// must default MaxHistory to 10 (never emit an invalid 0).
		_, err := scheduler.UpgradeV2(ctx, &pbSchedulerV2.SchedulerV2UpgradeV2Request{
			Name:  "app",
			Chart: "chart-v3",
		})
		require.NoError(t, err)

		wf, err := workflows.Get(ctx, "app", meta.GetOptions{})
		require.NoError(t, err)

		require.Equal(t, "chart-v3", wf.Spec.Chart.Name)
		require.NotNil(t, wf.Spec.Upgrade)
		require.NotNil(t, wf.Spec.Upgrade.MaxHistory)
		require.Equal(t, 10, *wf.Spec.Upgrade.MaxHistory)
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

func Test_Workflow_Permissions(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	// newImpl builds a sidecar whose SelfSubjectAccessReviews are answered with the given allow decision.
	newImpl := func(t *testing.T, allowed bool) *implementation {
		client := kclient.NewFakeClient()
		client.Kubernetes().(*kubernetesFake.Clientset).PrependReactor("create", "selfsubjectaccessreviews",
			func(action k8sTesting.Action) (bool, runtime.Object, error) {
				return true, &authorizationApi.SelfSubjectAccessReview{
					Status: authorizationApi.SubjectAccessReviewStatus{Allowed: allowed},
				}, nil
			})

		i, err := newInternal(client, nil, NewConfiguration().With(func(c Configuration) Configuration {
			c.Deployment = "test-deployment"
			c.Namespace = tests.FakeNamespace
			return c
		}))
		require.NoError(t, err)
		return i
	}

	t.Run("Denied when the ServiceAccount cannot manage workflows", func(t *testing.T) {
		err := newImpl(t, false).checkWorkflowPermissions(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not allowed")
		// The error lists the verbs the sidecar needs.
		for _, verb := range workflowRequiredVerbs {
			require.Contains(t, err.Error(), verb)
		}
	})

	t.Run("Allowed when the ServiceAccount can manage workflows", func(t *testing.T) {
		require.NoError(t, newImpl(t, true).checkWorkflowPermissions(ctx))
	})
}
