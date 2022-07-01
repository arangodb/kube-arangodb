//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func Test_PlanBuilderAppender_Recovery(t *testing.T) {
	t.Run("Recover", func(t *testing.T) {
		require.Len(t, recoverPlanAppender(testLogger, newPlanAppender(NewWithPlanBuilder(context.Background(), nil, api.DeploymentSpec{}, api.DeploymentStatus{}, nil), nil, nil)).
			Apply(func(_ context.Context, _ k8sutil.APIObject, _ api.DeploymentSpec, _ api.DeploymentStatus, _ PlanBuilderContext) api.Plan {
				panic("")
			}).
			Apply(func(_ context.Context, _ k8sutil.APIObject, _ api.DeploymentSpec, _ api.DeploymentStatus, _ PlanBuilderContext) api.Plan {
				panic("SomePanic")
			}).Plan(), 0)
	})
	t.Run("Recover with output", func(t *testing.T) {
		require.Len(t, recoverPlanAppender(testLogger, newPlanAppender(NewWithPlanBuilder(context.Background(), nil, api.DeploymentSpec{}, api.DeploymentStatus{}, nil), nil, nil)).
			Apply(func(_ context.Context, _ k8sutil.APIObject, _ api.DeploymentSpec, _ api.DeploymentStatus, _ PlanBuilderContext) api.Plan {
				return api.Plan{api.Action{}}
			}).
			ApplyIfEmpty(func(_ context.Context, _ k8sutil.APIObject, _ api.DeploymentSpec, _ api.DeploymentStatus, _ PlanBuilderContext) api.Plan {
				panic("SomePanic")
			}).
			ApplyIfEmpty(func(_ context.Context, _ k8sutil.APIObject, _ api.DeploymentSpec, _ api.DeploymentStatus, _ PlanBuilderContext) api.Plan {
				return api.Plan{api.Action{}, api.Action{}}
			}).Plan(), 1)
	})
	t.Run("Recover with multi", func(t *testing.T) {
		require.Len(t, recoverPlanAppender(testLogger, newPlanAppender(NewWithPlanBuilder(context.Background(), nil, api.DeploymentSpec{}, api.DeploymentStatus{}, nil), nil, nil)).
			Apply(func(_ context.Context, _ k8sutil.APIObject, _ api.DeploymentSpec, _ api.DeploymentStatus, _ PlanBuilderContext) api.Plan {
				return api.Plan{api.Action{}}
			}).
			ApplyIfEmpty(func(_ context.Context, _ k8sutil.APIObject, _ api.DeploymentSpec, _ api.DeploymentStatus, _ PlanBuilderContext) api.Plan {
				panic("SomePanic")
			}).
			Apply(func(_ context.Context, _ k8sutil.APIObject, _ api.DeploymentSpec, _ api.DeploymentStatus, _ PlanBuilderContext) api.Plan {
				return api.Plan{api.Action{}, api.Action{}}
			}).Plan(), 3)
	})
}
