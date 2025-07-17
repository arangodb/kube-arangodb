//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb-helper/go-helper/pkg/arangod/conn"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// newCompactMemberAction creates a new Action that implements the given
// planned CompactMember action.
func newCompactMemberAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionCompactMember{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionCompactMember implements an CompactMemberAction.
type actionCompactMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

func (a *actionCompactMember) Start(ctx context.Context) (bool, error) {
	m, g, ok := a.actionCtx.GetMemberStatusAndGroupByID(a.action.MemberID)
	if !ok {
		return false, errors.Errorf("expecting member to be present in list, but it is not")
	}

	switch g {
	case api.ServerGroupDBServers:
		dbc, err := a.actionCtx.GetServerAsyncClient(m.ID)
		if err != nil {
			return false, errors.Wrapf(err, "Unable to create client")
		}

		c := client.NewClient(dbc.Connection(), logger)

		if err := c.Compact(ctx, &client.CompactRequest{
			CompactBottomMostLevel: util.NewType(true),
			ChangeLevel:            util.NewType(true),
		}); err != nil {
			if id, ok := conn.IsAsyncJobInProgress(err); ok {
				a.actionCtx.Add(LocalJobID, id, true)

				return false, nil
			}

			return false, err
		}
	}

	return true, nil
}

func (a actionCompactMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	m, g, ok := a.actionCtx.GetMemberStatusAndGroupByID(a.action.MemberID)
	if !ok {
		return false, true, errors.Errorf("expecting member to be present in list, but it is not")
	}

	job, ok := a.actionCtx.Get(a.action, LocalJobID)
	if !ok {
		return false, true, errors.Errorf("Local Key is missing in action: %s", LocalJobID)
	}

	switch g {
	case api.ServerGroupDBServers:
		dbc, err := a.actionCtx.GetServerAsyncClient(m.ID)
		if err != nil {
			return false, false, errors.Wrapf(err, "Unable to create client")
		}

		c := client.NewClient(dbc.Connection(), logger)

		if err := c.Compact(conn.WithAsyncID(ctx, job), &client.CompactRequest{
			CompactBottomMostLevel: util.NewType(true),
			ChangeLevel:            util.NewType(true),
		}); err != nil {
			if _, ok := conn.IsAsyncJobInProgress(err); ok {
				return false, false, nil
			}

			if ok := conn.IsAsyncErrorNotFound(err); ok {
				// Job not found
				return false, true, err
			}

			return false, false, err
		}

		// Job Completed
		return true, false, nil
	}

	return true, false, nil
}
