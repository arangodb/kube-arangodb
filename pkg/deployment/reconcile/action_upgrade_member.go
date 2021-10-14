//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeUpgradeMember, newUpgradeMemberAction)
}

// newUpgradeMemberAction creates a new Action that implements the given
// planned UpgradeMember action.
func newUpgradeMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionUpgradeMember{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, upgradeMemberTimeout)

	return a
}

// actionUpgradeMember implements an UpgradeMember.
type actionUpgradeMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionUpgradeMember) Start(ctx context.Context) (bool, error) {
	log := a.log
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
	}
	// Set AutoUpgrade condition
	m.Conditions.Update(api.ConditionTypeAutoUpgrade, true, "Upgrading", "AutoUpgrade on first restart")

	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, errors.WithStack(err)
	}

	act := actionRotateMember{
		actionImpl: a.actionImpl,
	}

	return act.Start(ctx)
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionUpgradeMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Check that pod is removed
	log := a.log
	m, found := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !found {
		log.Error().Msg("No such member")
		return true, false, nil
	}

	if m.Phase == api.MemberPhaseRotating {
		act := actionRotateMember{
			actionImpl: a.actionImpl,
		}

		if _, abort, err := act.CheckProgress(ctx); err != nil || abort {
			return false, abort, err
		}

		return false, false, nil
	}

	isUpgrading := m.Phase == api.MemberPhaseUpgrading

	if isUpgrading {
		if m.Conditions.IsTrue(api.ConditionTypeTerminated) {
			if m.Conditions.IsTrue(api.ConditionTypeUpgradeFailed) {
				a.log.Error().Msgf("Upgrade of member failed")
			}
			// Invalidate plan
			m.Phase = ""
			m.Conditions.Remove(api.ConditionTypeTerminated)
			m.Conditions.Remove(api.ConditionTypeUpgradeFailed)

			if m.OldImage != nil {
				m.Image = m.OldImage.DeepCopy()
			}

			if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
				return false, true, nil
			}

			log.Error().Msgf("Upgrade failed")
			return false, true, nil
		}
	}

	log = log.With().
		Str("pod-name", m.PodName).
		Bool("is-upgrading", isUpgrading).Logger()

	act := actionWaitForMemberUp{
		actionImpl: a.actionImpl,
	}

	if ok, _, err := act.CheckProgress(ctx); err != nil {
		return false, false, errors.WithStack(err)
	} else if !ok {
		return false, false, nil
	}

	// Pod is now upgraded, update the member status
	m.Phase = api.MemberPhaseCreated
	m.RecentTerminations = nil // Since we're upgrading, we do not care about old terminations.
	m.CleanoutJobID = ""
	if !m.OldImage.Equal(m.Image) && isUpgrading {
		m.OldImage = m.Image.DeepCopy()
	}
	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, false, errors.WithStack(err)
	}
	return isUpgrading, false, nil
}
