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
	"net/http"

	driver "github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

// newWaitForMemberUpAction creates a new Action that implements the given
// planned WaitForMemberUp action.
func newWaitForMemberUpAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionWaitForMemberUp{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionWaitForMemberUp implements an WaitForMemberUp.
type actionWaitForMemberUp struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionWaitForMemberUp) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return ready, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionWaitForMemberUp) CheckProgress(ctx context.Context) (bool, bool, error) {
	member, ok := a.actionCtx.GetMemberStatusByID(a.MemberID())
	if !ok || member.Phase == api.MemberPhaseFailed {
		a.log.Debug("Member in failed phase")
		return true, false, nil
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	if a.action.Group.IsArangosync() {
		return a.checkProgressArangoSync(ctxChild)
	}
	switch a.actionCtx.GetMode() {
	case api.DeploymentModeSingle:
		return a.checkProgressSingle(ctxChild)
	case api.DeploymentModeActiveFailover:
		if a.action.Group == api.ServerGroupAgents {
			return a.checkProgressAgent()
		}
		return a.checkProgressSingleInActiveFailover(ctxChild)
	default:
		if a.action.Group == api.ServerGroupAgents {
			return a.checkProgressAgent()
		}
		return a.checkProgressCluster(ctx)
	}
}

// checkProgressSingle checks the progress of the action in the case
// of a single server.
func (a *actionWaitForMemberUp) checkProgressSingle(ctx context.Context) (bool, bool, error) {
	c, err := a.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		a.log.Err(err).Debug("Failed to create database client")
		return false, false, nil
	}
	if _, err := c.Version(ctx); err != nil {
		a.log.Err(err).Debug("Failed to get version")
		return false, false, nil
	}
	return true, false, nil
}

// checkProgressSingleInActiveFailover checks the progress of the action in the case
// of a single server as part of an active failover deployment.
func (a *actionWaitForMemberUp) checkProgressSingleInActiveFailover(ctx context.Context) (bool, bool, error) {
	c, err := a.actionCtx.GetMembersState().GetMemberClient(a.action.MemberID)
	if err != nil {
		a.log.Err(err).Debug("Failed to create database client")
		return false, false, nil
	}
	if _, err := c.Version(ctx); err != nil {
		a.log.Err(err).Debug("Failed to get version")
		return false, false, nil
	}
	return true, false, nil
}

// checkProgressAgent checks the progress of the action in the case
// of an agent.
func (a *actionWaitForMemberUp) checkProgressAgent() (bool, bool, error) {
	agencyHealth, ok := a.actionCtx.GetAgencyHealth()
	if !ok {
		a.log.Debug("Agency health fetch failed")
		return false, false, nil
	}
	if err := agencyHealth.Healthy(); err != nil {
		a.log.Err(err).Debug("Not all agents are ready")
		return false, false, nil
	}

	a.log.Debug("Agency is happy")

	return true, false, nil
}

// checkProgressCluster checks the progress of the action in the case
// of a cluster deployment (coordinator/dbserver).
func (a *actionWaitForMemberUp) checkProgressCluster(ctx context.Context) (bool, bool, error) {
	h, _ := a.actionCtx.GetMembersState().Health()
	if h.Error != nil {
		a.log.Err(h.Error).Debug("Cluster health is missing")
		return false, false, nil
	}
	sh, found := h.Members[driver.ServerID(a.action.MemberID)]
	if !found {
		a.log.Debug("Member not yet found in cluster health")
		return false, false, nil
	}
	if sh.Status != driver.ServerStatusGood {
		a.log.Str("status", string(sh.Status)).Debug("Member set status not yet good")
		return false, false, nil
	}

	// Wait for the member to become ready from a kubernetes point of view
	// otherwise the coordinators may be rotated to fast and thus none of them
	// is ready resulting in a short downtime
	m, found := a.actionCtx.GetMemberStatusByID(a.MemberID())
	if !found {
		a.log.Error("No such member")
		return false, true, nil
	}

	imageInfo, found := a.actionCtx.GetCurrentImageInfo()
	if !found {
		a.log.Info("Image not found")
		return false, false, nil
	}

	if resources.IsServerProgressAvailable(a.action.Group, imageInfo) {
		if status, err := a.getServerStatus(ctx); err == nil {
			progress, _ := status.GetProgress()
			a.actionCtx.SetProgress(progress)
		} else {
			a.log.Err(err).Warn("Failed to get server status to establish a progress")
		}
	}

	if !m.Conditions.IsTrue(api.ConditionTypeReady) {
		a.log.Debug("Member not yet ready")
		return false, false, nil
	}

	return true, false, nil
}

// checkProgressArangoSync checks the progress of the action in the case
// of a sync master / worker.
func (a *actionWaitForMemberUp) checkProgressArangoSync(ctx context.Context) (bool, bool, error) {
	c, err := a.actionCtx.GetMembersState().GetMemberSyncClient(a.action.MemberID)
	if err != nil {
		a.log.Err(err).Debug("Failed to create arangosync client")
		return false, false, nil
	}

	// When replication is in initial-sync state, then it can take a long time to be in running state.
	// This is the reason why Health of ArangoSync can not be checked here.
	if _, err := c.Version(ctx); err != nil {
		a.log.Err(err).Debug("Member is not ready yet")
		return false, false, nil
	}
	return true, false, nil
}

func (a actionWaitForMemberUp) getServerStatus(ctx context.Context) (client.ServerStatus, error) {
	cli, err := a.actionCtx.GetMembersState().GetMemberClient(a.action.MemberID)
	if err != nil {
		return client.ServerStatus{}, err
	}
	conn := cli.Connection()

	req, err := conn.NewRequest("GET", "_admin/status")
	if err != nil {
		return client.ServerStatus{}, err
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	resp, err := conn.Do(ctxChild, req)
	if err != nil {
		return client.ServerStatus{}, err
	}

	if err := resp.CheckStatus(http.StatusOK); err != nil {
		return client.ServerStatus{}, err
	}

	var result client.ServerStatus

	if err := resp.ParseBody("", &result); err != nil {
		return client.ServerStatus{}, err
	}

	return result, nil
}
