//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeMemberPhaseUpdate, newMemberPhaseUpdate)
}

const (
	ActionTypeMemberPhaseUpdatePhaseKey string = "phase"
)

func newMemberPhaseUpdate(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &memberPhaseUpdateAction{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type memberPhaseUpdateAction struct {
	actionImpl

	actionEmptyCheckProgress
}

func (a *memberPhaseUpdateAction) Start(ctx context.Context) (bool, error) {
	log := a.log
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, nil
	}

	phaseString, ok := a.action.Params[ActionTypeMemberPhaseUpdatePhaseKey]
	if !ok {
		log.Error().Msg("Phase not defined")
		return true, nil
	}

	p, ok := api.GetPhase(phaseString)
	if !ok {
		log.Error().Msgf("Phase %s unknown", p)
		return true, nil
	}

	if phase.Execute(&m, a.action, p) {
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, errors.WithStack(err)
		}
	}

	return true, nil
}
