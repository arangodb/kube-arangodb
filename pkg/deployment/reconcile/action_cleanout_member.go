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

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
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
	var jobID string
	ctx = driver.WithJobIDResponse(ctx, &jobID)
	if err := cluster.CleanOutServer(ctx, a.action.MemberID); err != nil {
		log.Debug().Err(err).Msg("Failed to cleanout member")
		return false, maskAny(err)
	}
	log.Debug().Str("job-id", jobID).Msg("Cleanout member started")
	// Update status
	m.Phase = api.MemberPhaseCleanOut
	m.CleanoutJobID = jobID
	if a.actionCtx.UpdateMember(m); err != nil {
		return false, maskAny(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionCleanoutMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	log := a.log
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		// We wanted to remove and it is already gone. All ok
		return true, false, nil
	}
	// do not try to clean out a pod that was not initialized
	if !m.IsInitialized {
		return true, false, nil
	}
	c, err := a.actionCtx.GetDatabaseClient(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create database client")
		return false, false, maskAny(err)
	}
	cluster, err := c.Cluster(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to access cluster")
		return false, false, maskAny(err)
	}
	cleanedOut, err := cluster.IsCleanedOut(ctx, a.action.MemberID)
	if err != nil {
		log.Debug().Err(err).Msg("IsCleanedOut failed")
		return false, false, maskAny(err)
	}
	if !cleanedOut {
		// We're not done yet, check job status
		log.Debug().Msg("IsCleanedOut returned false")

		c, err := a.actionCtx.GetDatabaseClient(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create database client")
			return false, false, maskAny(err)
		}
		agency, err := a.actionCtx.GetAgency(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create agency client")
			return false, false, maskAny(err)
		}
		jobStatus, err := arangod.CleanoutServerJobStatus(ctx, m.CleanoutJobID, c, agency)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to fetch cleanout job status")
			return false, false, maskAny(err)
		}
		if jobStatus.IsFailed() {
			log.Warn().Str("reason", jobStatus.Reason()).Msg("Cleanout Job failed. Aborting plan")
			// Revert cleanout state
			m.Phase = api.MemberPhaseCreated
			m.CleanoutJobID = ""
			if a.actionCtx.UpdateMember(m); err != nil {
				return false, false, maskAny(err)
			}
			return false, true, nil
		}
		return false, false, nil
	}
	// Cleanout completed
	if m.Conditions.Update(api.ConditionTypeCleanedOut, true, "CleanedOut", "") {
		if a.actionCtx.UpdateMember(m); err != nil {
			return false, false, maskAny(err)
		}
	}
	// Cleanout completed
	return true, false, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *actionCleanoutMember) Timeout() time.Duration {
	return cleanoutMemberTimeout
}

// Return the MemberID used / created in this action
func (a *actionCleanoutMember) MemberID() string {
	return a.action.MemberID
}
