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
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createTaskPlan creates a plan for an unprocessed, the oldest task
// If a plan already exists, the given plan is returned with false.
// Otherwise, the new plan is returned with a boolean true.
func (r *Reconciler) createTaskPlan(ctx context.Context, apiObject k8sutil.APIObject,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus,
	builderCtx PlanBuilderContext) (api.Plan, api.BackOff, bool) {
	if !currentPlan.IsEmpty() {
		// Plan already exists, complete that first
		return currentPlan, nil, false
	}

	var plan api.Plan
	task, err := r.context.GetNextTask(ctx)
	if err != nil {
		r.log.Error("Failed to get next task: %v", err)
		return plan, nil, false
	}

	if task == nil {
		return plan, nil, false
	}

	r.log.Info("Starting processing task: %s", task.Name)
	switch task.Spec.Type {
	case api.ArangoTaskPingType:
		plan = r.createPingPlan(task)
	default:
		r.log.Error("Unknown task type: %s", task.Spec.Type)
	}

	return plan, status.BackOff, true
}

func (r *Reconciler) createPingPlan(task *api.ArangoTask) api.Plan {
	return api.Plan{
		actions.NewClusterAction(api.ActionTypePing, "Pinging database server").SetTaskID(task.UID),
	}
}
