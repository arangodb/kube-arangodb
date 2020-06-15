//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util"

	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func init() {
	registerAction(api.ActionTypeWaitForMemberInSync, newWaitForMemberInSync)
}

// newWaitForMemberUpAction creates a new Action that implements the given
// planned WaitForShardInSync action.
func newWaitForMemberInSync(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionWaitForMemberInSync{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, waitForMemberUpTimeout)

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
func (a *actionWaitForMemberInSync) CheckProgress(ctx context.Context) (bool, bool, error) {
	member, ok := a.actionCtx.GetMemberStatusByID(a.MemberID())
	if !ok || member.Phase == api.MemberPhaseFailed {
		a.log.Debug().Msg("Member in failed phase")
		return true, false, nil
	}

	ready, err := a.check(ctx)
	if err != nil {
		return false, false, err
	}

	return ready, false, nil
}

func (a *actionWaitForMemberInSync) check(ctx context.Context) (bool, error) {
	spec := a.actionCtx.GetSpec()

	groupSpec := spec.GetServerGroupSpec(a.action.Group)

	if !util.BoolOrDefault(groupSpec.ExtendedRotationCheck, false) {
		return true, nil
	}

	switch spec.Mode.Get() {
	case api.DeploymentModeCluster:
		return a.checkCluster(ctx, spec, groupSpec)
	default:
		return true, nil
	}
}

func (a *actionWaitForMemberInSync) checkCluster(ctx context.Context, spec api.DeploymentSpec, groupSpec api.ServerGroupSpec) (bool, error) {
	if !a.actionCtx.GetShardSyncStatus() {
		a.log.Info().Str("mode", "cluster").Msgf("Shards are not in sync")
		return false, nil
	}

	return true, nil
}
