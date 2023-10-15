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

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

const (
	resignLeadershipJobID api.PlanLocalKey = "resignLeadershipJobID"
)

// newEnforceResignLeadershipAction creates a new Action that implements the given
// planned ResignLeadership action.
func newEnforceResignLeadershipAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionEnforceResignLeadership{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionEnforceResignLeadership implements an ResignLeadershipAction.
type actionEnforceResignLeadership struct {
	actionImpl
}

// Start performs the start of the ReasignLeadership process on DBServer.
func (a *actionEnforceResignLeadership) Start(ctx context.Context) (bool, error) {
	group := a.action.Group

	if a.actionCtx.GetSpec().Mode.Get() != api.DeploymentModeCluster {
		a.log.Debug("Resign only allowed in cluster mode")
		return true, nil
	}

	switch group {
	case api.ServerGroupDBServers:
		if agencyState, agencyOK := a.actionCtx.GetAgencyCache(); !agencyOK {
			a.log.Warn("AgencyCache is not ready")
			return false, nil
		} else if agencyState.Supervision.Maintenance.Exists() {
			// We are done, action cannot be handled on maintenance mode
			a.log.Warn("Maintenance is enabled, skipping action")
			return true, nil
		}

		return false, nil
	default:
		return true, nil
	}
}

// CheckProgress checks if the Job is completed, if not then start it. Repeat in case of error or if still a leader
func (a *actionEnforceResignLeadership) CheckProgress(ctx context.Context) (bool, bool, error) {
	group := a.action.Group
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, false, nil
	}

	if group != api.ServerGroupDBServers {
		// Only DBServers can use ResignLeadership job
		return true, false, nil
	}

	agencyState, agencyOK := a.actionCtx.GetAgencyCache()
	if !agencyOK {
		a.log.Error("Unable to get maintenance mode")
		return false, false, nil
	} else if agencyState.Supervision.Maintenance.Exists() {
		a.log.Warn("Maintenance is enabled, skipping action")
		// We are done, action cannot be handled on maintenance mode
		return true, false, nil
	} else if isServerRebooted(a.log, a.action, agencyState, driver.ServerID(m.ID)) {
		return true, false, nil
	}

	// Lets start resign job if required
	if j, ok := a.actionCtx.Get(a.action, resignLeadershipJobID); ok && j != "" && j != "N/A" {
		_, jobStatus := agencyState.Target.GetJob(state.JobID(j))
		switch jobStatus {
		case state.JobPhaseFailed:
			a.log.Error("Resign server job failed")
			// Remove key
			a.actionCtx.Add(resignLeadershipJobID, "N/A", true)
			return false, false, nil
		case state.JobPhaseFinished:
			a.log.Info("Job finished")
			// Remove key
			a.actionCtx.Add(resignLeadershipJobID, "N/A", true)
		case state.JobPhaseUnknown:
			a.log.Str("status", string(jobStatus)).Error("Resign server job unknown status")
			return false, false, nil
		default:
			return false, false, nil
		}

		a.actionCtx.Add(resignLeadershipJobID, "N/A", true)

		// Job is Finished, check if we are not a leader anymore
		if agencyState.PlanLeaderServers().Contains(state.Server(m.ID)) {
			// We are still a leader!
			if agencyState.PlanLeaderServersWithFailOver().Contains(state.Server(m.ID)) {
				// We need to retry
				a.log.Warn("DBServers is still a leader for shards")
				return false, false, nil
			}
			// Nothing to do as RF is set to 1
			a.log.Warn("DBServers is still a leader for shards, but ReplicationFactor is set to 1")
		}
		return true, false, nil
	}

	// Job not in progress, start it
	client, err := a.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		a.log.Err(err).Error("Unable to get client")
		return false, false, errors.WithStack(err)
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cluster, err := client.Cluster(ctxChild)
	if err != nil {
		a.log.Err(err).Error("Unable to get cluster client")
		return false, false, errors.WithStack(err)
	}

	var jobID string
	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	jobCtx := driver.WithJobIDResponse(ctxChild, &jobID)
	a.log.Debug("Temporary shutdown, resign leadership")
	if err := cluster.ResignServer(jobCtx, m.ID); err != nil {
		a.log.Err(err).Debug("Failed to resign server")
		return false, false, errors.WithStack(err)
	}

	a.actionCtx.Add(resignLeadershipJobID, jobID, true)

	return false, false, nil
}
