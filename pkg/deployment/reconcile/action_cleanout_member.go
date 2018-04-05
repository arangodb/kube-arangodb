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

// NewCleanOutMemberAction creates a new Action that implements the given
// planned CleanOutMember action.
func NewCleanOutMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionCleanoutMember{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// actionCleanoutMember implements an CleanOutMemberAction.
type actionCleanoutMember struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionCleanoutMember) Start(ctx context.Context) (bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		// We wanted to remove and it is already gone. All ok
		return true, nil
	}
	log := a.log
	c, err := a.actionCtx.GetDatabaseClient(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create member client")
		return false, maskAny(err)
	}
	cluster, err := c.Cluster(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to access cluster")
		return false, maskAny(err)
	}
	if err := cluster.CleanOutServer(ctx, a.action.MemberID); err != nil {
		log.Debug().Err(err).Msg("Failed to cleanout member")
		return false, maskAny(err)
	}
	// Update status
	m.Phase = api.MemberPhaseCleanOut
	if a.actionCtx.UpdateMember(m); err != nil {
		return false, maskAny(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionCleanoutMember) CheckProgress(ctx context.Context) (bool, error) {
	log := a.log
	c, err := a.actionCtx.GetDatabaseClient(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create member client")
		return false, maskAny(err)
	}
	cluster, err := c.Cluster(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to access cluster")
		return false, maskAny(err)
	}
	cleanedOut, err := cluster.IsCleanedOut(ctx, a.action.MemberID)
	if err != nil {
		return false, maskAny(err)
	}
	if !cleanedOut {
		// We're not done yet
		return false, nil
	}
	// Cleanout completed
	return true, nil
}
