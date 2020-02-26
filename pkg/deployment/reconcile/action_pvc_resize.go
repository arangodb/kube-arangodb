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

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

// NewRotateMemberAction creates a new Action that implements the given
// planned RotateMember action.
func NewPVCResizeAction(log zerolog.Logger, action api.Action, actionCtx ActionContext) Action {
	return &actionPVCResize{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
	}
}

// actionRotateMember implements an RotateMember.
type actionPVCResize struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionPVCResize) Start(ctx context.Context) (bool, error) {
	log := a.log
	group := a.action.Group
	groupSpec := a.actionCtx.GetSpec().GetServerGroupSpec(group)
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		log.Error().Msg("No such member")
		return true, nil
	}

	if m.PersistentVolumeClaimName == "" {
		// Nothing to do, PVC is empty
		return true, nil
	}

	pvc, err := a.actionCtx.GetPvc(m.PersistentVolumeClaimName)
	if err != nil {
		if errors.IsNotFound(err) {
			return true, nil
		}

		return false, err
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
				if err := a.actionCtx.UpdatePvc(pvc); err != nil {
					return false, err
				}

				return false, nil
			} else if cmp > 0 {
				log.Error().Str("server-group", group.AsRole()).Str("pvc-storage-size", volumeSize.String()).Str("requested-size", requestedSize.String()).
					Msg("Volume size should not shrink")
				a.actionCtx.CreateEvent(k8sutil.NewCannotShrinkVolumeEvent(a.actionCtx.GetAPIObject(), pvc.Name))
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
	log := a.log
	m, found := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !found {
		log.Error().Msg("No such member")
		return true, false, nil
	}

	pvc, err := a.actionCtx.GetPvc(m.PersistentVolumeClaimName)
	if err != nil {
		if errors.IsNotFound(err) {
			return true, false, nil
		}

		return false, true, err
	}

	pv, err := a.actionCtx.GetPv(pvc.Spec.VolumeName)
	if err != nil {
		if errors.IsNotFound(err) {
			return true, false, nil
		}

		return false, true, err
	}

	if requestedSize, ok := pvc.Spec.Resources.Requests[core.ResourceStorage]; ok {
		if volumeSize, ok := pv.Spec.Capacity[core.ResourceStorage]; ok {
			cmp := volumeSize.Cmp(requestedSize)
			if cmp >= 0 {
				return true, false, nil
			}
		}
	}

	return false, false, nil
}

// Timeout returns the amount of time after which this action will timeout.
func (a *actionPVCResize) Timeout() time.Duration {
	return pvcResizeTimeout
}

// Return the MemberID used / created in this action
func (a *actionPVCResize) MemberID() string {
	return a.action.MemberID
}
