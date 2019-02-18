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
	"github.com/arangodb/go-driver/agency"
	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

// NewWaitForMemberUpAction creates a new Action that implements the given
// planned WaitForMemberUp action.
func NewWaitForMemberUpAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionWaitForMemberUp{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// actionWaitForMemberUp implements an WaitForMemberUp.
type actionWaitForMemberUp struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionWaitForMemberUp) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	if err != nil {
		return false, maskAny(err)
	}
	return ready, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionWaitForMemberUp) CheckProgress(ctx context.Context) (bool, bool, error) {
	if a.action.Group.IsArangosync() {
		return a.checkProgressArangoSync(ctx)
	}
	switch a.actionCtx.GetMode() {
	case api.DeploymentModeSingle:
		return a.checkProgressSingle(ctx)
	case api.DeploymentModeActiveFailover:
		if a.action.Group == api.ServerGroupAgents {
			return a.checkProgressAgent(ctx)
		}
		return a.checkProgressSingleInActiveFailover(ctx)
	default:
		if a.action.Group == api.ServerGroupAgents {
			return a.checkProgressAgent(ctx)
		}
		return a.checkProgressCluster(ctx)
	}
}

// checkProgressSingle checks the progress of the action in the case
// of a single server.
func (a *actionWaitForMemberUp) checkProgressSingle(ctx context.Context) (bool, bool, error) {
	log := a.log
	c, err := a.actionCtx.GetDatabaseClient(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create database client")
		return false, false, maskAny(err)
	}
	if _, err := c.Version(ctx); err != nil {
		log.Debug().Err(err).Msg("Failed to get version")
		return false, false, maskAny(err)
	}
	return true, false, nil
}

// checkProgressSingleInActiveFailover checks the progress of the action in the case
// of a single server as part of an active failover deployment.
func (a *actionWaitForMemberUp) checkProgressSingleInActiveFailover(ctx context.Context) (bool, bool, error) {
	log := a.log
	c, err := a.actionCtx.GetDatabaseClient(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create database client")
		return false, false, maskAny(err)
	}
	if _, err := c.Version(ctx); err != nil {
		log.Debug().Err(err).Msg("Failed to get version")
		return false, false, maskAny(err)
	}
	if _, err := c.Databases(ctx); err != nil {
		log.Debug().Err(err).Msg("Failed to get databases")
		return false, false, maskAny(err)
	}
	return true, false, nil
}

// checkProgressAgent checks the progress of the action in the case
// of an agent.
func (a *actionWaitForMemberUp) checkProgressAgent(ctx context.Context) (bool, bool, error) {
	log := a.log
	clients, err := a.actionCtx.GetAgencyClients(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create agency clients")
		return false, false, maskAny(err)
	}

	if err := agency.AreAgentsHealthy(ctx, clients); err != nil {
		log.Debug().Err(err).Msg("Not all agents are ready")
		return false, false, nil
	}

	log.Debug().Msg("Agency is happy")

	return true, false, nil
}

// checkProgressCluster checks the progress of the action in the case
// of a cluster deployment (coordinator/dbserver).
func (a *actionWaitForMemberUp) checkProgressCluster(ctx context.Context) (bool, bool, error) {
	log := a.log
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
	h, err := cluster.Health(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get cluster health")
		return false, false, maskAny(err)
	}
	sh, found := h.Health[driver.ServerID(a.action.MemberID)]
	if !found {
		log.Debug().Msg("Member not yet found in cluster health")
		return false, false, nil
	}
	if sh.Status != driver.ServerStatusGood {
		log.Debug().Str("status", string(sh.Status)).Msg("Member set status not yet good")
		return false, false, nil
	}
	// Wait for the member to become ready from a kubernetes point of view
	// otherwise the coordinators may be rotated to fast and thus none of them
	// is ready resulting in a short downtime
	if m, found := a.actionCtx.GetMemberStatusByID(a.MemberID()); !found {
		log.Error().Msg("No such member")
		return false, true, nil
	} else if !m.Conditions.IsTrue(api.ConditionTypeReady) {
		log.Debug().Msg("Member not yet ready")
		return false, false, nil
	}
	return true, false, nil
}

// checkProgressArangoSync checks the progress of the action in the case
// of a sync master / worker.
func (a *actionWaitForMemberUp) checkProgressArangoSync(ctx context.Context) (bool, bool, error) {
	log := a.log
	c, err := a.actionCtx.GetSyncServerClient(ctx, a.action.Group, a.action.MemberID)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create arangosync client")
		return false, false, maskAny(err)
	}
	if err := c.Health(ctx); err != nil {
		log.Debug().Err(err).Msg("Health not ok yet")
		return false, false, maskAny(err)
	}
	return true, false, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *actionWaitForMemberUp) Timeout() time.Duration {
	return waitForMemberUpTimeout
}

// Return the MemberID used / created in this action
func (a *actionWaitForMemberUp) MemberID() string {
	return a.action.MemberID
}
