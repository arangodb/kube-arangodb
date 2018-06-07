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

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

// NewRemoveMemberAction creates a new Action that implements the given
// planned RemoveMember action.
func NewRemoveMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionRemoveMember{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// actionRemoveMember implements an RemoveMemberAction.
type actionRemoveMember struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
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
				return false, maskAny(errors.Wrapf(err, "Failed to remove server from cluster: %#v", err))
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
	return true, nil
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionRemoveMember) CheckProgress(ctx context.Context) (bool, error) {
	// Nothing todo
	return true, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *actionRemoveMember) Timeout() time.Duration {
	return removeMemberTimeout
}
