//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	ObsoleteClusterConditions = []api.ConditionType{
		api.ConditionTypeMaintenanceMode,
	}
)

func (r *Reconciler) cleanupConditions(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
	var p api.Plan

	for _, c := range ObsoleteClusterConditions {
		if _, ok := status.Conditions.Get(c); ok {
			p = append(p, shared.RemoveConditionActionV2("Cleanup Condition", c))
		}
	}

	return p
}
