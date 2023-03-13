//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

// newDrainMemberAction creates a new Action that implements the given
// planned DrainMember action.
func newDrainMemberAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionDrainMember{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionDrainMember implements an DrainMemberAction.
type actionDrainMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionDrainMember) Start(ctx context.Context) (bool, error) {
	if features.RestartWithTermination().Enabled() {
		return true, nil
	}

	if a.action.Group != a.actionCtx.GetSpec().Mode.Get().ServingGroup() {
		a.log.Debug("Preparation for shutdown required only on serving groups")
		return true, nil
	}

	if err := a.actionCtx.WithStatusUpdateErr(ctx, func(s *api.DeploymentStatus) (bool, error) {
		status, g, ok := s.Members.ElementByID(a.action.MemberID)
		if !ok {
			a.log.Info("can not set the condition because the member is gone already")
			return false, nil
		}

		if g != a.action.Group {
			a.log.Info("can not set the condition because of invalid groups")
			return false, nil
		}

		if status.Conditions.UpdateWithHash(api.ConditionTypeDrain, true, "Prepare member shutdown", "Prepare member shutdown", "") {
			if err := s.Members.Update(status, g); err != nil {
				return false, err
			}

			return true, nil
		}

		return false, nil
	}); err != nil {
		a.log.Err(err).Warn("unable to update status")
		return true, nil
	}

	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionDrainMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	if features.RestartWithTermination().Enabled() {
		return true, false, nil
	}

	s := a.actionCtx.GetStatus()

	status, _, ok := s.Members.ElementByID(a.action.MemberID)
	if !ok {
		a.log.Info("can not set the condition because the member is gone already")
		return true, false, nil
	}

	if status.Conditions.IsTrue(api.ConditionTypeServing) {
		return false, false, nil
	} else {
		return true, false, nil
	}
}
