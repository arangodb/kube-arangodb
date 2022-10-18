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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// newRotateMemberAction creates a new Action that implements the given
// planned RotateMember action.
func newPVCResizeAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionPVCResize{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

// actionRotateMember implements an RotateMember.
type actionPVCResize struct {
	// actionImpl implement timeout and member id functions
	actionImpl
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionPVCResize) Start(ctx context.Context) (bool, error) {
	group := a.action.Group
	groupSpec := a.actionCtx.GetSpec().GetServerGroupSpec(group)
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Error("No such member")
		return true, nil
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		return true, errors.Newf("Cluster is not ready")
	}

	if m.PersistentVolumeClaim.GetName() == "" {
		// Nothing to do, PVC is empty
		return true, nil
	}

	pvc, ok := cache.PersistentVolumeClaim().V1().GetSimple(m.PersistentVolumeClaim.GetName())
	if !ok {
		return true, nil
	}

	var res core.ResourceList
	if groupSpec.HasVolumeClaimTemplate() {
		res = groupSpec.GetVolumeClaimTemplate().Spec.Resources.Requests
	} else {
		res = groupSpec.Resources.Requests
	}

	if requestedSize, ok := res[core.ResourceStorage]; ok {
		if volumeSize, ok := pvc.Spec.Resources.Requests[core.ResourceStorage]; ok {
			cmp := volumeSize.Cmp(requestedSize)
			if cmp < 0 {
				pvc.Spec.Resources.Requests[core.ResourceStorage] = requestedSize
				nctx, c := globals.GetGlobals().Timeouts().Kubernetes().WithTimeout(ctx)
				defer c()

				if _, err := cache.Client().Kubernetes().CoreV1().PersistentVolumeClaims(cache.Namespace()).Update(nctx, pvc, meta.UpdateOptions{}); err != nil {
					return false, err
				}

				return false, nil
			}
		}
	}

	return true, nil
}

// CheckProgress checks the progress of the action.
// Returns: ready, abort, error.
func (a *actionPVCResize) CheckProgress(ctx context.Context) (bool, bool, error) {
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
