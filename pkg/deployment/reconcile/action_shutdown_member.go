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
)

const (
	shutdownTimeout = time.Second * 15
)

// NewShutdownMemberAction creates a new Action that implements the given
// planned ShutdownMember action.
func NewShutdownMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionShutdownMember{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// actionShutdownMember implements an ShutdownMemberAction.
type actionShutdownMember struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionShutdownMember) Start(ctx context.Context) (bool, error) {
	log := a.log
	group := a.action.Group
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, nil
	}
	if group.IsArangod() {
		// Invoke shutdown endpoint
		c, err := a.actionCtx.GetServerClient(ctx, group, a.action.MemberID)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create member client")
			return false, maskAny(err)
		}
		removeFromCluster := true
		log.Debug().Bool("removeFromCluster", removeFromCluster).Msg("Shutting down member")
		ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()
		if err := c.Shutdown(ctx, removeFromCluster); err != nil {
			// Shutdown failed. Let's check if we're already done
			if ready, _, err := a.CheckProgress(ctx); err == nil && ready {
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
	m.Phase = api.MemberPhaseShuttingDown
	if err := a.actionCtx.UpdateMember(m); err != nil {
		return false, maskAny(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionShutdownMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	m, found := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !found {
		// Member not long exists
		return true, false, nil
	}
	if m.Conditions.IsTrue(api.ConditionTypeTerminated) {
		// Shutdown completed
		return true, false, nil
	}
	// Member still not shutdown, retry soon
	return false, false, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *actionShutdownMember) Timeout() time.Duration {
	return shutdownMemberTimeout
}
