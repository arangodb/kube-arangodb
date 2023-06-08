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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// newWaitForMemberUpAction creates a new Action that implements the given
// planned WaitForShardInSync action.
func newWaitForMemberInSyncAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionWaitForMemberInSync{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionWaitForMemberInSync implements an WaitForShardInSync.
type actionWaitForMemberInSync struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionWaitForMemberInSync) Start(ctx context.Context) (bool, error) {
	ready, _, err := a.CheckProgress(ctx)
	return ready, err
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionWaitForMemberInSync) CheckProgress(_ context.Context) (bool, bool, error) {
	member, ok := a.actionCtx.GetMemberStatusByID(a.MemberID())
	if !ok || member.Phase == api.MemberPhaseFailed {
		a.log.Debug("Member in failed phase")
		return true, false, nil
	}

	ready, err := a.check()
	if err != nil {
		return false, false, err
	}

	return ready, false, nil
}

func (a *actionWaitForMemberInSync) check() (bool, error) {
	spec := a.actionCtx.GetSpec()

	groupSpec := spec.GetServerGroupSpec(a.action.Group)

	if !util.TypeOrDefault[bool](groupSpec.ExtendedRotationCheck, false) {
		return true, nil
	}

	switch spec.Mode.Get() {
	case api.DeploymentModeCluster:
		return a.checkCluster()
	default:
		return true, nil
	}
}

func (a *actionWaitForMemberInSync) checkCluster() (bool, error) {
	switch a.action.Group {
	case api.ServerGroupDBServers:
		agencyState, ok := a.actionCtx.GetAgencyCache()
		if !ok {
			a.log.Str("mode", "cluster").Str("member", a.MemberID()).Info("AgencyCache is missing")
			return false, nil
		}

		notInSyncShards := state.GetDBServerShardsNotInSync(agencyState, state.Server(a.MemberID()))

		if len(notInSyncShards) > 0 {
			a.log.Str("mode", "cluster").Str("member", a.MemberID()).Int("shard", len(notInSyncShards)).Info("DBServer contains not in sync shards")
			return false, nil
		}
	case api.ServerGroupAgents:
		agencyHealth, ok := a.actionCtx.GetAgencyHealth()
		if !ok {
			a.log.Str("mode", "cluster").Str("member", a.MemberID()).Info("AgencyHealth is missing")
			return false, nil
		}
		if err := agencyHealth.Healthy(); err != nil {
			a.log.Str("mode", "cluster").Str("member", a.MemberID()).Err(err).Info("Agency is not yet synchronized")
			return false, nil
		}
	}
	return true, nil
}
