//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

// newDehibernateMemberAction creates a new Action that implements the given
// planned DehibernateMember action.
func newDehibernateMemberAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionDehibernateMember{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionDehibernateMember implements a DehibernateMember. It moves a hibernated member back to the
// MemberPhaseNone phase, which is the operator's existing "recreate the pod" trigger, so the normal
// create path rebuilds the member and brings it back up.
type actionDehibernateMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionDehibernateMember) Start(ctx context.Context) (bool, error) {
	m, _, ok := a.actionCtx.GetMemberStatusAndGroupByID(a.action.MemberID)
	if !ok {
		return false, errors.Errorf("expecting member to be present in list, but it is not")
	}

	if !m.Phase.IsHibernated() {
		// Nothing to do, member is not hibernated.
		return true, nil
	}

	// Trigger the normal recreate path.
	m.Phase = api.MemberPhaseNone

	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, errors.WithStack(err)
	}

	return true, nil
}
