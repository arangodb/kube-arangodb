//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

func init() {
	registerAction(api.ActionTypeCleanOutMember, newCleanOutMemberAction)
}

// newCleanOutMemberAction creates a new Action that implements the given
// planned CleanOutMember action.
func newCleanOutMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionCleanoutMember{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, cleanoutMemberTimeout)

	return a
}

// actionCleanoutMember implements an CleanOutMemberAction.
type actionCleanoutMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
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

	ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
	c, err := a.actionCtx.GetDatabaseClient(ctxChild)
	cancel()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create member client")
		return false, errors.WithStack(err)
	}

	ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
	cluster, err := c.Cluster(ctxChild)
	cancel()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to access cluster")
		return false, errors.WithStack(err)
	}

	ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
	defer cancel()
	var jobID string
	ctxJobID := driver.WithJobIDResponse(ctxChild, &jobID)
	if err := cluster.CleanOutServer(ctxJobID, a.action.MemberID); err != nil {
		log.Debug().Err(err).Msg("Failed to cleanout member")
		return false, errors.WithStack(err)
	}
	log.Debug().Str("job-id", jobID).Msg("Cleanout member started")
	// Update status
	m.Phase = api.MemberPhaseCleanOut
	m.CleanoutJobID = jobID
	if a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, errors.WithStack(err)
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

	ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
	c, err := a.actionCtx.GetDatabaseClient(ctxChild)
	cancel()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create database client")
		return false, false, errors.WithStack(err)
	}

	ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
	cluster, err := c.Cluster(ctxChild)
	cancel()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to access cluster")
		return false, false, errors.WithStack(err)
	}

	ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
	cleanedOut, err := cluster.IsCleanedOut(ctxChild, a.action.MemberID)
	cancel()
	if err != nil {
		log.Debug().Err(err).Msg("IsCleanedOut failed")
		return false, false, errors.WithStack(err)
	}
	if !cleanedOut {
		// We're not done yet, check job status
		log.Debug().Msg("IsCleanedOut returned false")

		ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
		c, err := a.actionCtx.GetDatabaseClient(ctxChild)
		cancel()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create database client")
			return false, false, errors.WithStack(err)
		}

		ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
		agency, err := a.actionCtx.GetAgency(ctxChild)
		cancel()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create agency client")
			return false, false, errors.WithStack(err)
		}

		ctxChild, cancel = context.WithTimeout(ctx, arangod.GetRequestTimeout())
		jobStatus, err := arangod.CleanoutServerJobStatus(ctxChild, m.CleanoutJobID, c, agency)
		cancel()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to fetch cleanout job status")
			return false, false, errors.WithStack(err)
		}
		if jobStatus.IsFailed() {
			log.Warn().Str("reason", jobStatus.Reason()).Msg("Cleanout Job failed. Aborting plan")
			// Revert cleanout state
			m.Phase = api.MemberPhaseCreated
			m.CleanoutJobID = ""
			if a.actionCtx.UpdateMember(ctx, m); err != nil {
				return false, false, errors.WithStack(err)
			}
			return false, true, nil
		}
		return false, false, nil
	}
	// Cleanout completed
	if m.Conditions.Update(api.ConditionTypeCleanedOut, true, "CleanedOut", "") {
		if a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, false, errors.WithStack(err)
		}
	}
	// Cleanout completed
	return true, false, nil
}
