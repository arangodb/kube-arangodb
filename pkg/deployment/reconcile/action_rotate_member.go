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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/rs/zerolog"
)

// NewRotateMemberAction creates a new Action that implements the given
// planned RotateMember action.
func NewRotateMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionRotateMember{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// actionRotateMember implements an RotateMember.
type actionRotateMember struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionRotateMember) Start(ctx context.Context) (bool, error) {
	log := a.log
	group := a.action.Group
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
	}
	if group.IsArangod() {
		// Invoke shutdown endpoint
		c, err := a.actionCtx.GetServerClient(ctx, group, a.action.MemberID)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create member client")
			return false, maskAny(err)
		}
		removeFromCluster := false
		log.Debug().Bool("removeFromCluster", removeFromCluster).Msg("Shutting down member")
		ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()
		if err := c.Shutdown(ctx, removeFromCluster); err != nil {
			// Shutdown failed. Let's check if we're already done
			if ready, err := a.CheckProgress(ctx); err == nil && ready {
				// We're done
				return true, nil
			}
			log.Debug().Err(err).Msg("Failed to shutdown member")
			return false, maskAny(err)
		}
	} else if group.IsArangosync() {
		// Terminate pod
		if err := a.actionCtx.DeletePod(m.PodName); err != nil {
			return false, maskAny(err)
		}
	}
	// Update status
	m.State = api.MemberStateRotating
	if err := a.actionCtx.UpdateMember(m); err != nil {
		return false, maskAny(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionRotateMember) CheckProgress(ctx context.Context) (bool, error) {
	// Check that pod is removed
	log := a.log
	m, found := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !found {
		log.Error().Msg("No such member")
		return true, nil
	}
	if !m.Conditions.IsTrue(api.ConditionTypeTerminated) {
		// Pod is not yet terminated
		return false, nil
	}
	// Pod is terminated, we can now remove it
	if err := a.actionCtx.DeletePod(m.PodName); err != nil {
		return false, maskAny(err)
	}
	// Pod is now gone, update the member status
	m.State = api.MemberStateNone
	if err := a.actionCtx.UpdateMember(m); err != nil {
		return false, maskAny(err)
	}
	return true, nil
}
