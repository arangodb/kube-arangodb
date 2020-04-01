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
// Author Ewout Prangsma
//

package reconcile

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

func init() {
	registerAction(api.ActionTypeRemoveMember, newRemoveMemberAction)
}

// newRemoveMemberAction creates a new Action that implements the given
// planned RemoveMember action.
func newRemoveMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionRemoveMember{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, removeMemberTimeout)

	return a
}

// actionRemoveMember implements an RemoveMemberAction.
type actionRemoveMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionRemoveMember) Start(ctx context.Context) (bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		// We wanted to remove and it is already gone. All ok
		return true, nil
	}
	// For safety, remove from cluster
	if a.action.Group == api.ServerGroupCoordinators || a.action.Group == api.ServerGroupDBServers {
		client, err := a.actionCtx.GetDatabaseClient(ctx)
		if err != nil {
			return false, maskAny(err)
		}
		if err := arangod.RemoveServerFromCluster(ctx, client.Connection(), driver.ServerID(m.ID)); err != nil {
			if !driver.IsNotFound(err) && !driver.IsPreconditionFailed(err) {
				a.log.Err(err).Str("member-id", m.ID).Msgf("Failed to remove server from cluster")
				// ignore this error, maybe all coordinators are failed and no connction to cluster is possible
			} else if driver.IsPreconditionFailed(err) {
				health, err := a.actionCtx.GetDeploymentHealth()
				if err != nil {
					return false, maskAny(errors.Wrapf(err, "failed to get cluster health"))
				}
				// We don't care if not found
				if record, ok := health.Health[driver.ServerID(m.ID)]; ok {

					// Check if the pod is terminating
					if m.Conditions.IsTrue(api.ConditionTypeTerminating) {

						if record.Status != driver.ServerStatusFailed {
							return false, maskAny(fmt.Errorf("can not remove server from cluster. Not yet terminated. Retry later"))
						}

						a.log.Debug().Msg("dbserver has shut down")
					}
				}
			} else {
				a.log.Warn().Msgf("ignoring error: %s", err.Error())
			}
		}
	}
	// Remove the pod (if any)
	if err := a.actionCtx.DeletePod(m.PodName); err != nil {
		return false, maskAny(err)
	}
	// Remove the pvc (if any)
	if m.PersistentVolumeClaimName != "" {
		if err := a.actionCtx.DeletePvc(m.PersistentVolumeClaimName); err != nil {
			return false, maskAny(err)
		}
	}
	// Remove member
	if err := a.actionCtx.RemoveMemberByID(a.action.MemberID); err != nil {
		return false, maskAny(err)
	}
	// Check that member has been removed
	if _, found := a.actionCtx.GetMemberStatusByID(a.action.MemberID); found {
		return false, maskAny(fmt.Errorf("Member %s still exists", a.action.MemberID))
	}
	return true, nil
}
