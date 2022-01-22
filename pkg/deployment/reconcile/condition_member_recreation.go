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
	"strings"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/rs/zerolog"
)

func createMemberRecreationConditionsPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	var p api.Plan

	for _, m := range status.Members.AsList() {
		resp, recreate := EvaluateMemberRecreationCondition(ctx, log, apiObject, spec, status, m.Group, m.Member, cachedStatus, context)

		if !recreate {
			if _, ok := m.Member.Conditions.Get(api.MemberReplacementRequired); ok {
				// Unset condition
				p = append(p, removeMemberConditionActionV2("Member replacement not required", api.MemberReplacementRequired, m.Group, m.Member.ID))
			}
		} else {
			if c, ok := m.Member.Conditions.Get(api.MemberReplacementRequired); !ok || !c.IsTrue() || c.Message != resp {
				// Update condition
				p = append(p, updateMemberConditionActionV2("Member replacement required", api.MemberReplacementRequired, m.Group, m.Member.ID, true, "Member replacement required", resp, ""))
			}
		}
	}

	return p
}

type MemberRecreationConditionEvaluator func(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	group api.ServerGroup, member api.MemberStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) (string, bool)

func EvaluateMemberRecreationCondition(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	group api.ServerGroup, member api.MemberStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext, evaluators ...MemberRecreationConditionEvaluator) (string, bool) {
	args := make([]string, 0, len(evaluators))

	for _, e := range evaluators {
		if s, ok := e(ctx, log, apiObject, spec, status, group, member, cachedStatus, context); ok {
			args = append(args, s)
		}
	}

	return strings.Join(args, ", "), len(args) > 0
}
