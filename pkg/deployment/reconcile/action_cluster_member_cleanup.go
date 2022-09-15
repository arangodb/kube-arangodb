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

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

// newClusterMemberCleanupAction creates a new Action that implements the given
// planned ClusterMemberCleanup action.
func newClusterMemberCleanupAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionClusterMemberCleanup{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionClusterMemberCleanup implements an ClusterMemberCleanup.
type actionClusterMemberCleanup struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionClusterMemberCleanup) Start(ctx context.Context) (bool, error) {
	if err := a.start(ctx); err != nil {
		a.log.Err(err).Warn("Unable to clean cluster member")
	}

	return true, nil
}

func (a *actionClusterMemberCleanup) start(ctx context.Context) error {
	id := driver.ServerID(a.MemberID())

	c, err := a.actionCtx.GetMembersState().State().GetDatabaseClient()
	if err != nil {
		return err
	}

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	cluster, err := c.Cluster(ctxChild)
	if err != nil {
		return err
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	health, err := cluster.Health(ctxChild)
	if err != nil {
		return err
	}

	if _, ok := health.Health[id]; !ok {
		return nil
	}

	return globals.GetGlobalTimeouts().ArangoD().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return cluster.RemoveServer(ctxChild, id)
	})
}
