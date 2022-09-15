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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/member"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	actionTypeMemberPhaseUpdatePhaseKey string = "phase"
)

func newMemberPhaseUpdateAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionMemberPhaseUpdate{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionMemberPhaseUpdate struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *actionMemberPhaseUpdate) Start(ctx context.Context) (bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, nil
	}

	phaseString, ok := a.action.Params[actionTypeMemberPhaseUpdatePhaseKey]
	if !ok {
		a.log.Error("Phase not defined")
		return true, nil
	}

	p, ok := api.GetPhase(phaseString)
	if !ok {
		a.log.Error("Phase %s unknown", p)
		return true, nil
	}

	if member.GetPhaseExecutor().Execute(a.actionCtx.GetAPIObject(), a.actionCtx.GetSpec(), a.action.Group, &m, a.action, p) {
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, errors.WithStack(err)
		}
	}

	return true, nil
}
