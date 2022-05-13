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

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
)

// createRotateMemberPlan creates a plan to rotate (stop-recreate-start) an existing
// member.
func createRotateMemberPlan(log zerolog.Logger, member api.MemberStatus,
	group api.ServerGroup, spec api.DeploymentSpec, reason string) api.Plan {
	log.Debug().
		Str("id", member.ID).
		Str("role", group.AsRole()).
		Str("reason", reason).
		Msg("Creating rotation plan")
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
		actions.NewAction(api.ActionTypeWaitForMemberUp, group, member),
		actions.NewAction(api.ActionTypeWaitForMemberInSync, group, member),
	)

	return plan
}

func emptyPlanBuilder(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	return nil
}

func removeConditionActionV2(actionReason string, conditionType api.ConditionType) api.Action {
	return actions.NewClusterAction(api.ActionTypeSetConditionV2, actionReason).
		AddParam(setConditionActionV2KeyAction, string(conditionType)).
		AddParam(setConditionActionV2KeyType, setConditionActionV2KeyTypeRemove)
}

//nolint:unparam
func updateConditionActionV2(actionReason string, conditionType api.ConditionType, status bool, reason, message, hash string) api.Action {
	statusBool := core.ConditionTrue
	if !status {
		statusBool = core.ConditionFalse
	}

	return actions.NewClusterAction(api.ActionTypeSetConditionV2, actionReason).
		AddParam(setConditionActionV2KeyAction, string(conditionType)).
		AddParam(setConditionActionV2KeyType, setConditionActionV2KeyTypeAdd).
		AddParam(setConditionActionV2KeyStatus, string(statusBool)).
		AddParam(setConditionActionV2KeyReason, reason).
		AddParam(setConditionActionV2KeyMessage, message).
		AddParam(setConditionActionV2KeyHash, hash)
}

func removeMemberConditionActionV2(actionReason string, conditionType api.ConditionType, group api.ServerGroup, member string) api.Action {
	return actions.NewAction(api.ActionTypeSetMemberConditionV2, group, withPredefinedMember(member), actionReason).
		AddParam(setConditionActionV2KeyAction, string(conditionType)).
		AddParam(setConditionActionV2KeyType, setConditionActionV2KeyTypeRemove)
}

func updateMemberConditionActionV2(actionReason string, conditionType api.ConditionType, group api.ServerGroup, member string, status bool, reason, message, hash string) api.Action {
	statusBool := core.ConditionTrue
	if !status {
		statusBool = core.ConditionFalse
	}

	return actions.NewAction(api.ActionTypeSetMemberConditionV2, group, withPredefinedMember(member), actionReason).
		AddParam(setConditionActionV2KeyAction, string(conditionType)).
		AddParam(setConditionActionV2KeyType, setConditionActionV2KeyTypeAdd).
		AddParam(setConditionActionV2KeyStatus, string(statusBool)).
		AddParam(setConditionActionV2KeyReason, reason).
		AddParam(setConditionActionV2KeyMessage, message).
		AddParam(setConditionActionV2KeyHash, hash)
}
