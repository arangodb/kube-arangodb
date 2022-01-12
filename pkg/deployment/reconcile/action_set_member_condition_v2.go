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

	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func init() {
	registerAction(api.ActionTypeSetMemberConditionV2, setMemberConditionV2)
}

func setMemberConditionV2(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionSetMemberConditionV2{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, defaultTimeout)

	return a
}

type actionSetMemberConditionV2 struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	actionEmptyCheckProgress
}

// Start starts the action for changing conditions on the provided member.
func (a actionSetMemberConditionV2) Start(ctx context.Context) (bool, error) {
	at, ok := a.action.Params[setConditionActionV2KeyType]
	if !ok {
		a.log.Info().Msgf("key %s is missing in action definition", setConditionActionV2KeyType)
		return true, nil
	}

	aa, ok := a.action.Params[setConditionActionV2KeyAction]
	if !ok {
		a.log.Info().Msgf("key %s is missing in action definition", setConditionActionV2KeyAction)
		return true, nil
	}

	switch at {
	case setConditionActionV2KeyTypeAdd:
		ah := a.action.Params[setConditionActionV2KeyHash]
		am := a.action.Params[setConditionActionV2KeyMessage]
		ar := a.action.Params[setConditionActionV2KeyReason]
		as := a.action.Params[setConditionActionV2KeyStatus] == string(core.ConditionTrue)

		if err := a.actionCtx.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
			m, _, ok := s.Members.ElementByID(a.action.MemberID)
			if !ok {
				a.log.Info().Msg("can not set the condition because the member is gone already")
				return false, nil
			}

			return m.Conditions.UpdateWithHash(api.ConditionType(aa), as, ar, am, ah), nil
		}); err != nil {
			a.log.Warn().Err(err).Msgf("unable to update status")
			return true, nil
		}
	case setConditionActionV2KeyTypeRemove:
		if err := a.actionCtx.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
			m, _, ok := s.Members.ElementByID(a.action.MemberID)
			if !ok {
				a.log.Info().Msg("can not set the condition because the member is gone already")
				return false, nil
			}

			return m.Conditions.Remove(api.ConditionType(aa)), nil
		}); err != nil {
			a.log.Warn().Err(err).Msgf("unable to update status")
			return true, nil
		}
	default:
		a.log.Info().Msgf("unknown type %s", at)
		return true, nil
	}
	return true, nil
}
