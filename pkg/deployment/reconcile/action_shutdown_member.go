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
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// newShutdownMemberAction creates a new Action that implements the given
// planned ShutdownMember action.
func newShutdownMemberAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionShutdownMember{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionShutdownMember implements an ShutdownMemberAction.
type actionShutdownMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionShutdownMember) Start(ctx context.Context) (bool, error) {
	shutdown, m, ok := getShutdownHelper(a.actionImpl)
	if !ok {
		return true, nil
	}

	if ready, err := shutdown.Start(ctx); err != nil {
		return false, err
	} else if ready {
		return true, nil
	}

	// Update status
	m.Phase = api.MemberPhaseShuttingDown
	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, errors.WithStack(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionShutdownMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	shutdown, _, ok := getShutdownHelper(a.actionImpl)
	if !ok {
		return true, false, nil
	}

	if ready, abort, err := shutdown.CheckProgress(ctx); err != nil {
		return false, abort, err
	} else {
		return ready, false, nil
	}
}
