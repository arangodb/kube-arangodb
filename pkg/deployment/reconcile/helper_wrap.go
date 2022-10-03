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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

func withMaintenance(plan ...api.Action) api.Plan {
	if !features.Maintenance().Enabled() {
		return plan
	}

	return withMaintenanceStart(plan...).After(actions.NewClusterAction(api.ActionTypeDisableMaintenance, "Disable maintenance after actions"))
}
func withMaintenanceStart(plan ...api.Action) api.Plan {
	if !features.Maintenance().Enabled() {
		return plan
	}

	return api.AsPlan(plan).Before(
		actions.NewClusterAction(api.ActionTypeEnableMaintenance, "Enable maintenance before actions"))
}

func withMemberMaintenance(group api.ServerGroup, member api.MemberStatus, reason string, plan api.Plan) api.Plan {
	if member.Image == nil {
		return plan
	}

	if group != api.ServerGroupDBServers {
		return plan
	}

	if !features.Version310().ImageSupported(member.Image) {
		return plan
	}

	return plan.Wrap(actions.NewAction(api.ActionTypeEnableMemberMaintenance, group, member, reason),
		actions.NewAction(api.ActionTypeDisableMemberMaintenance, group, member, reason))
}

func withResignLeadership(group api.ServerGroup, member api.MemberStatus, reason string, plan api.Plan) api.Plan {
	if member.Image == nil {
		return plan
	}

	return api.AsPlan(plan).Before(actions.NewAction(api.ActionTypeResignLeadership, group, member, reason))
}

func cleanOutMember(group api.ServerGroup, m api.MemberStatus) api.Plan {
	var plan api.Plan

	if group == api.ServerGroupDBServers {
		plan = append(plan,
			actions.NewAction(api.ActionTypeCleanOutMember, group, m),
		)
	}
	plan = append(plan,
		actions.NewAction(api.ActionTypeKillMemberPod, group, m),
		actions.NewAction(api.ActionTypeShutdownMember, group, m),
		actions.NewAction(api.ActionTypeRemoveMember, group, m),
	)

	return plan
}
