//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package shared

import (
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
)

const (
	SetConditionActionV2KeyTypeAdd    string = "add"
	SetConditionActionV2KeyTypeRemove string = "remove"

	SetConditionActionV2KeyType    string = "type"
	SetConditionActionV2KeyAction  string = "action"
	SetConditionActionV2KeyStatus  string = "status"
	SetConditionActionV2KeyReason  string = "reason"
	SetConditionActionV2KeyMessage string = "message"
	SetConditionActionV2KeyHash    string = "hash"
)

func RemoveConditionActionV2(actionReason string, conditionType api.ConditionType) api.Action {
	return actions.NewClusterAction(api.ActionTypeSetConditionV2, actionReason).
		AddParam(SetConditionActionV2KeyAction, string(conditionType)).
		AddParam(SetConditionActionV2KeyType, SetConditionActionV2KeyTypeRemove)
}

//nolint:unparam
func UpdateConditionActionV2(actionReason string, conditionType api.ConditionType, status bool, reason, message, hash string) api.Action {
	statusBool := core.ConditionTrue
	if !status {
		statusBool = core.ConditionFalse
	}

	return actions.NewClusterAction(api.ActionTypeSetConditionV2, actionReason).
		AddParam(SetConditionActionV2KeyAction, string(conditionType)).
		AddParam(SetConditionActionV2KeyType, SetConditionActionV2KeyTypeAdd).
		AddParam(SetConditionActionV2KeyStatus, string(statusBool)).
		AddParam(SetConditionActionV2KeyReason, reason).
		AddParam(SetConditionActionV2KeyMessage, message).
		AddParam(SetConditionActionV2KeyHash, hash)
}

func RemoveMemberConditionActionV2(actionReason string, conditionType api.ConditionType, group api.ServerGroup, member string) api.Action {
	return actions.NewAction(api.ActionTypeSetMemberConditionV2, group, WithPredefinedMember(member), actionReason).
		AddParam(SetConditionActionV2KeyAction, string(conditionType)).
		AddParam(SetConditionActionV2KeyType, SetConditionActionV2KeyTypeRemove)
}

//nolint:unparam
func UpdateMemberConditionActionV2(actionReason string, conditionType api.ConditionType, group api.ServerGroup, member string, status bool, reason, message, hash string) api.Action {
	statusBool := core.ConditionTrue
	if !status {
		statusBool = core.ConditionFalse
	}

	return actions.NewAction(api.ActionTypeSetMemberConditionV2, group, WithPredefinedMember(member), actionReason).
		AddParam(SetConditionActionV2KeyAction, string(conditionType)).
		AddParam(SetConditionActionV2KeyType, SetConditionActionV2KeyTypeAdd).
		AddParam(SetConditionActionV2KeyStatus, string(statusBool)).
		AddParam(SetConditionActionV2KeyReason, reason).
		AddParam(SetConditionActionV2KeyMessage, message).
		AddParam(SetConditionActionV2KeyHash, hash)
}
