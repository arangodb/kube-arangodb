//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NewAddMemberAction creates a new Action that implements the given
// planned AddMember action.
func NewAddMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionAddMember{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// actionAddMember implements an AddMemberAction.
type actionAddMember struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionAddMember) Start(ctx context.Context) (bool, error) {
	if err := a.actionCtx.CreateMember(a.action.Group, a.action.MemberID); err != nil {
		log.Debug().Err(err).Msg("Failed to create member")
		return false, maskAny(err)
	}
	return true, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionAddMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Nothing todo
	return true, false, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *actionAddMember) Timeout() time.Duration {
	return addMemberTimeout
}
