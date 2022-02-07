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

	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func init() {
	registerAction(api.ActionTypeRotateMember, newRotateMemberAction)
}

// newRotateMemberAction creates a new Action that implements the given
// planned RotateMember action.
func newRotateMemberAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	a := &actionRotateMember{}

	a.actionImpl = newActionImplDefRef(log, action, actionCtx, rotateMemberTimeout)

	return a
}

// actionRotateMember implements an RotateMember.
type actionRotateMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionRotateMember) Start(ctx context.Context) (bool, error) {
	shutdown, m, ok := getShutdownHelper(&a.action, a.actionCtx, a.log)
	if !ok {
		return true, nil
	}

	if ready, err := shutdown.Start(ctx); err != nil {
		return false, err
	} else if ready {
		return true, nil
	}

	// Update status
	m.Phase = api.MemberPhaseRotating

	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, errors.WithStack(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionRotateMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Check that pod is removed
	log := a.log
	shutdown, m, ok := getShutdownHelper(&a.action, a.actionCtx, a.log)
	if !ok {
		return true, false, nil
	}

	if ready, abort, err := shutdown.CheckProgress(ctx); err != nil {
		return false, abort, err
	} else if !ready {
		return false, false, nil
	}

	// Pod is terminated, we can now remove it
	if err := a.actionCtx.DeletePod(ctx, m.PodName, meta.DeleteOptions{}); err != nil {
		if !k8sutil.IsNotFound(err) {
			log.Error().Err(err).Msg("Unable to delete pod")
			return false, false, nil
		}
	}
	// Pod is now gone, update the member status
	m.Phase = api.MemberPhaseNone
	m.RecentTerminations = nil // Since we're rotating, we do not care about old terminations.
	m.CleanoutJobID = ""
	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, false, errors.WithStack(err)
	}
	return true, false, nil
}

func (a *actionRotateMember) StartFailureGracePeriod() time.Duration {
	return 30 * time.Second
}
