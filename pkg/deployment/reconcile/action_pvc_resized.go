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

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// newRotateMemberAction creates a new Action that implements the given
// planned RotateMember action.
func newPVCResizedAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionPVCResized{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionRotateMember implements an RotateMember.
type actionPVCResized struct {
	// actionImpl implement timeout and member id functions
	actionImpl

	// actionEmptyStart empty start function
	actionEmptyStart
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionPVCResized) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Check that pod is removed
	m, found := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !found {
		a.log.Error("No such member")
		return true, false, nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		a.log.Warn("Cluster is not ready")
		return false, false, nil
	}

	pvc, ok := cache.PersistentVolumeClaim().V1().GetSimple(m.PersistentVolumeClaim.GetName())
	if !ok {
		return true, false, nil
	}

	// If we are pending for FS to be resized - we need to proceed with mounting of PVC
	if k8sutil.IsPersistentVolumeClaimFileSystemResizePending(pvc) {
		return true, false, nil
	}

	if requestedSize, ok := pvc.Spec.Resources.Requests[core.ResourceStorage]; ok {
		if volumeSize, ok := pvc.Status.Capacity[core.ResourceStorage]; ok {
			cmp := volumeSize.Cmp(requestedSize)
			if cmp >= 0 {
				return true, false, nil
			}
		}
	}

	return false, false, nil
}
