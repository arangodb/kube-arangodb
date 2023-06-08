//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	driver "github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

// newCleanOutMemberAction creates a new Action that implements the given
// planned CleanOutMember action.
func newCleanOutMemberAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionCleanOutMember{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionCleanOutMember implements an CleanOutMemberAction.
type actionCleanOutMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionCleanOutMember) Start(ctx context.Context) (bool, error) {
	if a.action.Group != api.ServerGroupDBServers {
		// Proceed only on DBServers
		return true, nil
	}

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		// We wanted to remove and it is already gone. All ok
		return true, nil
	}

	c, err := a.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		a.log.Err(err).Debug("Failed to create member client")
		return false, errors.WithStack(err)
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cluster, err := c.Cluster(ctxChild)
	if err != nil {
		a.log.Err(err).Debug("Failed to access cluster")
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
		a.log.Err(err).Debug("Failed to cleanout member")
		return false, errors.WithStack(err)
	}
	a.log.Str("job-id", jobID).Debug("Cleanout member started")
	// Update status
	m.Phase = api.MemberPhaseCleanOut
	m.CleanoutJobID = jobID
	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, errors.WithStack(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionCleanOutMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		// We wanted to remove and it is already gone. All ok
		return true, false, nil
	}
	// do not try to clean out a pod that was not initialized
	if !m.IsInitialized {
		return true, false, nil
	}

	if m.Phase == api.MemberPhaseCreated {
		// Restart occurred
		m.Phase = api.MemberPhaseCleanOut
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, false, errors.WithStack(err)
		}
	}

	cache, ok := a.actionCtx.GetAgencyCache()
	if !ok {
		a.log.Debug("AgencyCache is not ready")
		return false, false, nil
	}

	if !cache.Target.CleanedServers.Contains(state.Server(a.action.MemberID)) {
		// We're not done yet, check job status
		a.log.Debug("IsCleanedOut returned false")

		details, jobStatus := cache.Target.GetJob(state.JobID(m.CleanoutJobID))
		if jobStatus == state.JobPhaseFailed {
			a.log.Str("reason", details.Reason).Warn("Cleanout Job failed. Aborting plan")
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

	if cache.PlanServers().Contains(state.Server(m.ID)) {
		// Something is wrong, servers is CleanedOut but still exists in the Plan
		a.actionCtx.CreateOperatorEngineOpsAlertEvent("DBServer %s still exists in Plan after CleanOut", m.ID)
		return false, true, nil
	}

	// Cleanout completed
	return true, false, nil
}
