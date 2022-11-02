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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

// newRotateStartMemberAction creates a new Action that implements the given
// planned RotateStartMember action.
func newRotateStartMemberAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRotateStartMember{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionRotateStartMember implements an RotateStartMember.
type actionRotateStartMember struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionRotateStartMember) Start(ctx context.Context) (bool, error) {
	shutdown, m, ok := getShutdownHelper(a.actionImpl)
	if !ok {
		return true, nil
	}

	if ready, err := shutdown.Start(ctx); err != nil {
		return false, err
	} else if ready {
		return true, nil
	}

	// Update status
	m.Phase = api.MemberPhaseRotateStart

	if err := a.actionCtx.UpdateMember(ctx, m); err != nil {
		return false, errors.WithStack(err)
	}
	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionRotateStartMember) CheckProgress(ctx context.Context) (bool, bool, error) {
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

	// Pod is terminated, we can now remove it
	if err := cache.Client().Kubernetes().CoreV1().Pods(cache.Namespace()).Delete(ctx, m.Pod.GetName(), meta.DeleteOptions{}); err != nil {
		if !kerrors.IsNotFound(err) {
			a.log.Err(err).Error("Unable to delete pod")
			return false, false, nil
		}
	}

	return true, false, nil
}
