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

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

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
	if a.action.Group != api.ServerGroupDBServers {
		// Proceed only on DBServers
		return true, nil
	}

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		// We wanted to remove and it is already gone. All ok
		return true, nil
	}
	log := a.log

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := a.actionCtx.GetDatabaseClient(ctxChild)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create member client")
		return false, errors.WithStack(err)
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cluster, err := c.Cluster(ctxChild)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to access cluster")
		return false, errors.WithStack(err)
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	var jobID string
	ctxJobID := driver.WithJobIDResponse(ctxChild, &jobID)
	if err := cluster.CleanOutServer(ctxJobID, a.action.MemberID); err != nil {
		if driver.IsNotFound(err) {
			// Member not found, it could be that it never connected to the cluster
			return true, nil
		}
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

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := a.actionCtx.GetDatabaseClient(ctxChild)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create database client")
		return false, false, nil
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cluster, err := c.Cluster(ctxChild)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to access cluster")
		return false, false, nil
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cleanedOut, err := cluster.IsCleanedOut(ctxChild, a.action.MemberID)
	if err != nil {
		log.Debug().Err(err).Msg("IsCleanedOut failed")
		return false, false, nil
	}
	if !cleanedOut {
		// We're not done yet, check job status
		log.Debug().Msg("IsCleanedOut returned false")

		ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		c, err := a.actionCtx.GetDatabaseClient(ctxChild)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create database client")
			return false, false, nil
		}

		ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		agency, err := a.actionCtx.GetAgency(ctxChild)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create agency client")
			return false, false, nil
		}

		ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		jobStatus, err := arangod.CleanoutServerJobStatus(ctxChild, m.CleanoutJobID, c, agency)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to fetch cleanout job status")
			return false, false, nil
		}
		if jobStatus.IsFailed() {
			log.Warn().Str("reason", jobStatus.Reason()).Msg("Cleanout Job failed. Aborting plan")
			// Revert cleanout state
			m.Phase = api.MemberPhaseCreated
			m.CleanoutJobID = ""
			if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
				return false, false, errors.WithStack(err)
			}
			return false, true, nil
		}
		return false, false, nil
	}
	// Cleanout completed
	if m.Conditions.Update(api.ConditionTypeCleanedOut, true, "CleanedOut", "") {
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, false, errors.WithStack(err)
		}
	}

	// Cleanout completed
	return true, false, nil
}
