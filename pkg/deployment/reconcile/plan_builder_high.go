//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/rs/zerolog"
)

func (d *Reconciler) CreateHighPlan(ctx context.Context, cachedStatus inspectorInterface.Inspector) (error, bool) {
	// Create plan
	apiObject := d.context.GetAPIObject()
	spec := d.context.GetSpec()
	status, lastVersion := d.context.GetStatus()
	builderCtx := newPlanBuilderContext(d.context)
	newPlan, changed := createHighPlan(ctx, d.log, apiObject, status.HighPriorityPlan, spec, status, cachedStatus, builderCtx)

	// If not change, we're done
	if !changed {
		return nil, false
	}

	// Save plan
	if len(newPlan) == 0 {
		// Nothing to do
		return nil, false
	}

	// Send events
	for id := len(status.Plan); id < len(newPlan); id++ {
		action := newPlan[id]
		d.context.CreateEvent(k8sutil.NewPlanAppendEvent(apiObject, action.Type.String(), action.Group.AsRole(), action.MemberID, action.Reason))
	}

	status.HighPriorityPlan = newPlan

	if err := d.context.UpdateStatus(ctx, status, lastVersion); err != nil {
		return errors.WithStack(err), false
	}
	return nil, true
}

// createHighPlan considers the given specification & status and creates a plan to get the status in line with the specification.
// If a plan already exists, the given plan is returned with false.
// Otherwise the new plan is returned with a boolean true.
func createHighPlan(ctx context.Context, log zerolog.Logger, apiObject k8sutil.APIObject,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus, cachedStatus inspectorInterface.Inspector,
	builderCtx PlanBuilderContext) (api.Plan, bool) {
	if !currentPlan.IsEmpty() {
		// Plan already exists, complete that first
		return currentPlan, false
	}

	// Check for various scenario's
	var plan api.Plan

	pb := NewWithPlanBuilder(ctx, log, apiObject, spec, status, cachedStatus, builderCtx)

	if plan.IsEmpty() {
		plan = pb.Apply(updateMemberPhasePlan)
	}

	if plan.IsEmpty() {
		plan = pb.Apply(createCleanOutPlan)
	}

	// Return plan
	return plan, true
}

// updateMemberPhasePlan creates plan to update member phase
func updateMemberPhasePlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var plan api.Plan

	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			if m.Phase == api.MemberPhaseNone {
				plan = append(plan,
					api.NewAction(api.ActionTypeMemberRIDUpdate, group, m.ID, "Regenerate member RID"),
					api.NewAction(api.ActionTypeMemberPhaseUpdate, group, m.ID,
						"Move to Pending phase").AddParam(ActionTypeMemberPhaseUpdatePhaseKey, api.MemberPhasePending.String()))
			}
		}

		return nil
	})

	return plan
}
