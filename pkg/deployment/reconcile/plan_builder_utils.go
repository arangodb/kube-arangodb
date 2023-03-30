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
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createRotateMemberPlan creates a plan to rotate (stop-recreate-start) an existing
// member.
func (r *Reconciler) createRotateMemberPlan(member api.MemberStatus,
	group api.ServerGroup, spec api.DeploymentSpec, reason string) api.Plan {
	r.log.
		Str("id", member.ID).
		Str("role", group.AsRole()).
		Str("reason", reason).
		Debug("Creating rotation plan")
	return createRotateMemberPlanWithAction(member, group, api.ActionTypeRotateMember, spec, reason)
}

// createRotateMemberPlanWithAction creates a plan to rotate (stop-<action>>-start) an existing
// member.
func createRotateMemberPlanWithAction(member api.MemberStatus,
	group api.ServerGroup, action api.ActionType, spec api.DeploymentSpec, reason string) api.Plan {

	var plan = api.Plan{
		actions.NewAction(api.ActionTypeCleanTLSKeyfileCertificate, group, member, "Remove server keyfile and enforce renewal/recreation"),
	}
	plan = withSecureWrap(member, group, spec, plan...)

	plan = plan.After(
		actions.NewAction(api.ActionTypeKillMemberPod, group, member, reason),
		actions.NewAction(action, group, member, reason),
		actions.NewAction(api.ActionTypeCleanMemberService, group, member, "Remove server service and enforce renewal/recreation"),
	)

	plan = withWaitForMember(plan, group, member)

	plan = withMemberMaintenance(group, member, "Enable member maintenance", plan)

	return plan
}

func (r *Reconciler) emptyPlanBuilder(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	return nil
}
