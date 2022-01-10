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

	"github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func init() {
	registerAction(api.ActionTypeResignLeadership, newResignLeadershipAction)
}

// newResignLeadershipAction creates a new Action that implements the given
// planned ResignLeadership action.
func newResignLeadershipAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionResignLeadership{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, shutdownMemberTimeout)

	return a
}

// actionResignLeadership implements an ResignLeadershipAction.
type actionResignLeadership struct {
	actionImpl
}

// Start performs the start of the ReasignLeadership process on DBServer.
func (a *actionResignLeadership) Start(ctx context.Context) (bool, error) {
	log := a.log
	group := a.action.Group
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, nil
	}

	if a.actionCtx.GetSpec().Mode.Get() != api.DeploymentModeCluster {
		log.Debug().Msg("Resign only allowed in cluster mode")
		return true, nil
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	client, err := a.actionCtx.GetDatabaseClient(ctxChild)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to get client")
		return true, errors.WithStack(err)
	}

	switch group {
	case api.ServerGroupDBServers:
		if agencyState, agencyOK := a.actionCtx.GetAgencyCache(); !agencyOK {
			log.Warn().Err(err).Msgf("Maintenance is enabled, skipping action")
			return true, errors.WithStack(err)
		} else if agencyState.Supervision.Maintenance {
			// We are done, action cannot be handled on maintenance mode
			log.Warn().Msgf("Maintenance is enabled, skipping action")
			return true, nil
		}

		ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		cluster, err := client.Cluster(ctxChild)
		if err != nil {
			log.Error().Err(err).Msgf("Unable to get cluster client")
			return true, errors.WithStack(err)
		}

		var jobID string
		ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		jobCtx := driver.WithJobIDResponse(ctxChild, &jobID)
		log.Debug().Msg("Temporary shutdown, resign leadership")
		if err := cluster.ResignServer(jobCtx, m.ID); err != nil {
			log.Debug().Err(err).Msg("Failed to resign server")
			return true, errors.WithStack(err)
		}

		m.CleanoutJobID = jobID

		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return true, errors.WithStack(err)
		}

		return false, nil
	default:
		return true, nil
	}
}

// CheckProgress checks if Job is completed.
func (a *actionResignLeadership) CheckProgress(ctx context.Context) (bool, bool, error) {
	log := a.log

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, false, nil
	}

	if agencyState, agencyOK := a.actionCtx.GetAgencyCache(); !agencyOK {
		log.Error().Msgf("Unable to get maintenance mode")
		return false, false, nil
	} else if agencyState.Supervision.Maintenance {
		log.Warn().Msgf("Maintenance is enabled, skipping action")
		// We are done, action cannot be handled on maintenance mode
		m.CleanoutJobID = ""
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, false, errors.WithStack(err)
		}
		return true, false, nil
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	agency, err := a.actionCtx.GetAgency(ctxChild)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create agency client")
		return false, false, nil
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := a.actionCtx.GetDatabaseClient(ctxChild)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create member client")
		return false, false, nil
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	jobStatus, err := arangod.CleanoutServerJobStatus(ctxChild, m.CleanoutJobID, c, agency)
	if err != nil {
		if driver.IsNotFound(err) {
			log.Debug().Err(err).Msg("Job not found, but proceeding")
			return true, false, nil
		}
		log.Debug().Err(err).Msg("Failed to fetch job status")
		return false, false, errors.WithStack(err)
	}

	if jobStatus.IsFailed() {
		m.CleanoutJobID = ""
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, false, errors.WithStack(err)
		}
		log.Error().Msg("Resign server job failed")
		return true, false, nil
	}

	if jobStatus.IsFinished() {
		m.CleanoutJobID = ""
		if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
			return false, false, errors.WithStack(err)
		}
		return true, false, nil
	}

	return false, false, nil
}
