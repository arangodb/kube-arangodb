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
// Author Adam Janikowski
//

package reconcile

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

// NewRotateStopMemberAction creates a new Action that implements the given
// planned RotateStopMember action.
func NewRotateStopMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionRotateStopMember{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// actionRotateStopMember implements an RotateStopMember.
type actionRotateStopMember struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionRotateStopMember) Start(ctx context.Context) (bool, error) {
	log := a.log
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
	}

	m.Phase = api.MemberPhaseNone
	m.RecentTerminations = nil // Since we're rotating, we do not care about old terminations.
	m.CleanoutJobID = ""
	if err := a.actionCtx.UpdateMember(m); err != nil {
		return false, maskAny(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionRotateStopMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	return true, false, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *actionRotateStopMember) Timeout() time.Duration {
	return rotateMemberTimeout
}

// Return the MemberID used / created in this action
func (a *actionRotateStopMember) MemberID() string {
	return a.action.MemberID
}
