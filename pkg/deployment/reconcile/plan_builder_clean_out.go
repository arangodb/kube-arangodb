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
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createCleanOutPlan creates clean out action if the server is cleaned out and the operator is not aware of it.
func (r *Reconciler) createCleanOutPlan(ctx context.Context, _ k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, planCtx PlanBuilderContext) api.Plan {

	if spec.GetMode() != api.DeploymentModeCluster {
		return nil
	}

	var plan api.Plan

	cache, ok := planCtx.GetAgencyCache()
	if !ok {
		r.log.Debug("AgencyCache is not ready")
		return nil
	}

	for _, m := range status.Members.AsListInGroup(api.ServerGroupDBServers) {
		cleanedStatus := m.Member.Conditions.IsTrue(api.ConditionTypeCleanedOut)
		cleanedAgency := cache.Target.CleanedServers.Contains(state.Server(m.Member.ID))

		if cleanedStatus != cleanedAgency {
			if cleanedAgency {
				plan = append(plan, shared.UpdateMemberConditionActionV2("DBServer cleaned", api.ConditionTypeCleanedOut, m.Group, m.Member.ID, true, "DBServer cleaned", "DBServer cleaned", ""))
			} else {
				plan = append(plan, shared.RemoveMemberConditionActionV2("DBServer is not cleaned", api.ConditionTypeCleanedOut, m.Group, m.Member.ID))
			}
		}
	}

	return plan
}
