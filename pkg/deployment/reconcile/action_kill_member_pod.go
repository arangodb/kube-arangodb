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
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/handlers/utils"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// newKillMemberPodAction creates a new Action that implements the given
// planned KillMemberPod action.
func newKillMemberPodAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionKillMemberPod{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionKillMemberPod implements an KillMemberPod.
type actionKillMemberPod struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionKillMemberPod) Start(ctx context.Context) (bool, error) {
	if !features.GracefulShutdown().Enabled() {
		return true, nil
	}

	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		return true, errors.Newf("Client is not ready")
	}

	if ifPodUIDMismatch(m, a.action, cache) {
		a.log.Error("Member UID is changed")
		return true, nil
	}

	if err := cache.Client().Kubernetes().CoreV1().Pods(cache.Namespace()).Delete(ctx, m.Pod.GetName(), meta.DeleteOptions{}); err != nil {
		a.log.Err(err).Error("Unable to kill pod")
		return true, nil
	}

	return false, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionKillMemberPod) CheckProgress(ctx context.Context) (bool, bool, error) {
	if !features.GracefulShutdown().Enabled() {
		return true, false, nil
	}
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, false, nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		return false, false, errors.Newf("Client is not ready")
	}

	p, ok := cache.Pod().V1().GetSimple(m.Pod.GetName())
	if !ok {
		a.log.Error("No such member")
		return true, false, nil
	}

	l := utils.StringList(p.Finalizers)

	if !l.Has(constants.FinalizerPodGracefulShutdown) {
		return true, false, nil
	}

	if l.Has(constants.FinalizerDelayPodTermination) {
		return false, false, nil
	}

	return true, false, nil
}
