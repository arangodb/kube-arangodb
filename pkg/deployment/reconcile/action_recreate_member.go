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
	"fmt"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	"time"

	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

// NewRecreateMemberAction creates a new Action that implements the given
// planned RecreateMember action.
func NewRecreateMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionRecreateMember{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// actionRecreateMember implements an RecreateMemberAction.
type actionRecreateMember struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionRecreateMember) Start(ctx context.Context) (bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		return false, fmt.Errorf("expecting member to be present in list, but it is not")
	}

	if m.Phase == api.MemberPhaseFailed {
		// Change cluster phase to ensure it wont be removed
		m.Phase = api.MemberPhaseNone
	}

	_, err := a.actionCtx.GetPvc(m.PersistentVolumeClaimName)
	if err != nil {
		if kubeErrors.IsNotFound(err) {
			return false, fmt.Errorf("PVC is missing %s. DBServer wont be recreated without old PV", m.PersistentVolumeClaimName)
		}

		return false, maskAny(err)
	}

	if err = a.actionCtx.UpdateMember(m); err != nil {
		return false, maskAny(err)
	}

	return true, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionRecreateMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Nothing todo
	return true, false, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *actionRecreateMember) Timeout() time.Duration {
	return recreateMemberTimeout
}

// Return the MemberID used / created in this action
func (a *actionRecreateMember) MemberID() string {
	return a.action.MemberID
}
