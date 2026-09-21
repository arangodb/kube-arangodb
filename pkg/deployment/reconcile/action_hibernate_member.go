//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

// newHibernateMemberAction creates a new Action that implements the given
// planned HibernateMember action.
func newHibernateMemberAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionHibernateMember{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionHibernateMember implements a HibernateMember. It gracefully shuts the member down (reusing the
// same shutdown helper as RotateStartMember) and parks it in the Hibernated phase, so it stays down
// until the deployment is dehibernated. A hibernated member is never recreated by EnsurePods (which
// only acts on Pending members).
type actionHibernateMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionHibernateMember) Start(ctx context.Context) (bool, error) {
	shutdown, m, ok := getShutdownHelper(a.actionImpl)
	if !ok {
		return true, nil
	}

	// Begin the graceful shutdown of the member. The member is only parked in the Hibernated phase
	// once the shutdown has completed (see hibernate).
	ready, err := shutdown.Start(ctx)
	if err != nil {
		return false, err
	}

	if ready {
		// Member is already down, park it in the Hibernated phase now.
		return true, a.hibernate(ctx, m)
	}

	// Shutdown is in progress, wait for the pod to terminate in CheckProgress.
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionHibernateMember) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Check that pod is removed
	shutdown, m, ok := getShutdownHelper(a.actionImpl)
	if !ok {
		return true, false, nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		a.log.Warn("Cluster is not ready")
		return false, false, nil
	}

	if ready, abort, err := shutdown.CheckProgress(ctx); err != nil {
		return false, abort, err
	} else if !ready {
		return false, false, nil
	}

	// Shutdown has completed. Park the member in the Hibernated phase before removing the pod so that
	// the now-missing pod is treated as an intended shutdown and not recreated.
	if err := a.hibernate(ctx, m); err != nil {
		return false, false, err
	}

	// Pod is terminated, we can now remove it
	if err := cache.Client().Kubernetes().CoreV1().Pods(cache.Namespace()).Delete(ctx, m.Pod.GetName(), meta.DeleteOptions{}); err != nil {
		if !kerrors.IsNotFound(err) {
			a.log.Err(err).Error("Unable to delete pod")
			return false, false, nil
		}
	}

	return true, false, nil
}

// hibernate parks the member in the Hibernated phase if it is not already there.
func (a *actionHibernateMember) hibernate(ctx context.Context, m api.MemberStatus) error {
	if m.Phase == api.MemberPhaseHibernated {
		return nil
	}

	m.Phase = api.MemberPhaseHibernated

	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
