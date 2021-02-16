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
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	registerAction(api.ActionTypeAddMember, newAddMemberAction)
}

// newAddMemberAction creates a new Action that implements the given
// planned AddMember action.
func newAddMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionAddMember{}

	a.actionImpl = newBaseActionImpl(log, action, actionCtx, func(deploymentSpec api.DeploymentSpec) time.Duration {
		return deploymentSpec.Timeouts.Get().AddMember.Get(addMemberTimeout)
	}, &a.newMemberID)

	return a
}

// actionAddMember implements an AddMemberAction.
type actionAddMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress

	newMemberID string
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionAddMember) Start(ctx context.Context) (bool, error) {
	newID, err := a.actionCtx.CreateMember(a.action.Group, a.action.MemberID)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create member")
		return false, errors.WithStack(err)
	}
	a.newMemberID = newID

	if _, ok := a.action.Params[api.ActionTypeWaitForMemberUp.String()]; ok {
		a.actionCtx.WithStatusUpdate(func(s *api.DeploymentStatus) bool {
			s.Plan = append(s.Plan, api.NewAction(api.ActionTypeWaitForMemberInSync, a.action.Group, newID, "Wait for member in sync after creation"))
			return true
		})
	}

	if _, ok := a.action.Params[api.ActionTypeWaitForMemberInSync.String()]; ok {
		a.actionCtx.WithStatusUpdate(func(s *api.DeploymentStatus) bool {
			s.Plan = append(s.Plan, api.NewAction(api.ActionTypeWaitForMemberInSync, a.action.Group, newID, "Wait for member in sync after creation"))
			return true
		})
	}

	return true, nil
}
