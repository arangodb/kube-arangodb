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

package actions

import api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

func NewAction(actionType api.ActionType, group api.ServerGroup, member api.MemberStatus, reason ...string) api.Action {
	return actionWrap(api.NewAction(actionType, group, member.ID, reason...), &member, actionWrapMemberUID)
}

func NewClusterAction(actionType api.ActionType, reason ...string) api.Action {
	m := api.MemberStatus{}
	return NewAction(actionType, api.ServerGroupUnknown, m, reason...)
}

func NewActionBuilderWrap(group api.ServerGroup, member api.MemberStatus) api.ActionBuilder {
	return &actionBuilderWrap{
		group:  group,
		member: member,
	}
}

type actionBuilderWrap struct {
	group  api.ServerGroup
	member api.MemberStatus
}

func (a actionBuilderWrap) NewAction(actionType api.ActionType, reason ...string) api.Action {
	return NewAction(actionType, a.group, a.member, reason...)
}

func (a actionBuilderWrap) Group() api.ServerGroup {
	return a.group
}

func (a actionBuilderWrap) MemberID() string {
	return a.member.ID
}

func actionWrap(a api.Action, member *api.MemberStatus, wrap ...actionWrapper) api.Action {
	for _, w := range wrap {
		a = w(a, member)
	}

	return a
}

func actionWrapMemberUID(a api.Action, member *api.MemberStatus) api.Action {
	switch a.Type {
	case api.ActionTypeShutdownMember, api.ActionTypeKillMemberPod, api.ActionTypeRotateStartMember, api.ActionTypeUpgradeMember:
		if q := member.Pod.GetUID(); q != "" {
			return a.AddParam(api.ParamPodUID, string(q))
		}
		return a
	default:
		return a
	}
}

type actionWrapper func(a api.Action, member *api.MemberStatus) api.Action
