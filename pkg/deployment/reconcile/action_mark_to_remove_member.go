//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeMarkToRemoveMember, newMarkToRemoveMemberAction)
}

func newMarkToRemoveMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionMarkToRemove{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, addMemberTimeout)

	return a
}

type actionMarkToRemove struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress
}

func (a *actionMarkToRemove) Start(ctx context.Context) (bool, error) {
	if a.action.Group != api.ServerGroupDBServers && a.action.Group != api.ServerGroupAgents && a.action.Group != api.ServerGroupCoordinators {
		return true, nil
	}

	return true, a.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		member, group, ok := s.Members.ElementByID(a.action.MemberID)
		if !ok {
			return false
		}

		if group != a.action.Group {
			return false
		}

		if !member.Conditions.Update(api.ConditionTypeMarkedToRemove, true, "Member marked to be removed", "") {
			return false
		}

		if err := s.Members.Update(member, group); err != nil {
			a.log.Warn().Err(err).Str("Member", member.ID).Msgf("Unable to update member")
			return false
		}

		return true
	})
}
