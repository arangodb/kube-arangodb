//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
)

// newRemoveMemberPVCAction creates a new Action that implements the given
// planned RemoveMemberPVC action.
func newRemoveMemberPVCAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionRemoveMemberPVC{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionRemoveMemberPVC implements an RemoveMemberPVCAction.
type actionRemoveMemberPVC struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyCheckProgress implement check progress with empty implementation
	actionEmptyCheckProgress
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionRemoveMemberPVC) Start(ctx context.Context) (bool, error) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		return true, nil
	}

	pvcUID, ok := a.action.GetParam("pvc")
	if !ok {
		return true, errors.Newf("PVC UID Parameter is missing")
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		return true, errors.Newf("Cluster is not ready")
	}

	agencyCache, ok := a.actionCtx.GetAgencyCache()
	if !ok {
		return true, errors.Newf("Agency is not ready")
	}

	if agencyCache.PlanLeaderServers().Contains(state.Server(m.ID)) {
		return true, errors.Newf("Server is still used in cluster")
	}

	// We are safe to remove PVC
	if pvcStatus := m.PersistentVolumeClaim; pvcStatus != nil {
		if n := pvcStatus.GetName(); n != "" {
			nctx, c := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
			defer c()
			err := cache.PersistentVolumeClaimsModInterface().V1().Delete(nctx, n, meta.DeleteOptions{
				Preconditions: meta.NewUIDPreconditions(pvcUID),
			})

			if err != nil {
				if apiErrors.IsNotFound(err) {
					// PVC is already gone
					return true, nil
				}

				if apiErrors.IsConflict(err) {
					// UID Changed, all fine
					return true, nil
				}

				return true, err
			}
		}
	}

	return true, nil
}
