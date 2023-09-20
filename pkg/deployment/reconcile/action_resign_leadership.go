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
	"strconv"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

const (
	actionResignLeadershipRebootID api.PlanLocalKey = "rebootID"
)

// newResignLeadershipAction creates a new Action that implements the given
// planned ResignLeadership action.
func newResignLeadershipAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionResignLeadership{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionResignLeadership implements an ResignLeadershipAction.
type actionResignLeadership struct {
	actionImpl
}

// Start performs the start of the ReasignLeadership process on DBServer.
func (a *actionResignLeadership) Start(ctx context.Context) (bool, error) {
	group := a.action.Group
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, nil
	}

	if a.actionCtx.GetSpec().Mode.Get() != api.DeploymentModeCluster {
		a.log.Debug("Resign only allowed in cluster mode")
		return true, nil
	}

	client, err := a.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		a.log.Err(err).Error("Unable to get client")
		return false, errors.WithStack(err)
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

		ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		cluster, err := client.Cluster(ctxChild)
		if err != nil {
			a.log.Err(err).Error("Unable to get cluster client")
			return false, errors.WithStack(err)
		}

		var jobID string
		ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		jobCtx := driver.WithJobIDResponse(ctxChild, &jobID)
		a.log.Debug("Temporary shutdown, resign leadership")
		if err := cluster.ResignServer(jobCtx, m.ID); err != nil {
			a.log.Err(err).Debug("Failed to resign server")
			return false, errors.WithStack(err)
		}

		m.CleanoutJobID = jobID

		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, errors.WithStack(err)
		}

		return false, nil
	default:
		return true, nil
	}
}

// CheckProgress checks if Job is completed.
func (a *actionResignLeadership) CheckProgress(ctx context.Context) (bool, bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, false, nil
	}

	agencyState, agencyOK := a.actionCtx.GetAgencyCache()
	if !agencyOK {
		a.log.Error("Unable to get maintenance mode")
		return false, false, nil
	} else if agencyState.Supervision.Maintenance.Exists() {
		a.log.Warn("Maintenance is enabled, skipping action")
		// We are done, action cannot be handled on maintenance mode
		m.CleanoutJobID = ""
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, false, errors.WithStack(err)
		}
		return true, false, nil
	} else if a.isServerRebooted(agencyState, driver.ServerID(m.ID)) {
		return true, false, nil
	}

	_, jobStatus := agencyState.Target.GetJob(state.JobID(m.CleanoutJobID))
	switch jobStatus {
	case state.JobPhaseFailed:
		m.CleanoutJobID = ""
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, false, errors.WithStack(err)
		}
		a.log.Error("Resign server job failed")
		return true, false, nil
	case state.JobPhaseFinished:
		m.CleanoutJobID = ""
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, false, errors.WithStack(err)
		}
		return true, false, nil
	case state.JobPhaseUnknown:
		a.log.Debug("Job not found, but proceeding")
		return true, false, nil
	}
	return false, false, nil
}

// isServerRebooted returns true when a given server ID was rebooted during resignation of leadership.
func (a *actionResignLeadership) isServerRebooted(agencyState state.State, serverID driver.ServerID) bool {
	rebootID, ok := agencyState.GetRebootID(serverID)
	if !ok {
		return false
	}

	v, ok := a.action.Params[actionResignLeadershipRebootID.String()]
	if !ok {
		a.log.Warn("missing reboot ID in action's locals")
		return false
	}

	r, err := strconv.Atoi(v)
	if err != nil {
		a.log.Err(err).Warn("reboot ID '%s' supposed to be a number", v)
		return false
	}

	if rebootID <= r {
		// Server has not been restarted.
		return false
	}

	a.log.Warn("resign leadership aborted because rebootID has changed from %d to %d", r, rebootID)
	return true
}
