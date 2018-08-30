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

package resources

import (
	"context"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	inspectedPVCCounter     = metrics.MustRegisterCounter("deployment", "inspected_ppvcs", "Number of PVCs inspections")
	maxPVCInspectorInterval = util.Interval(time.Hour) // Maximum time between PVC inspection (if nothing else happens)
)

// InspectPVCs lists all PVCs that belong to the given deployment and updates
// the member status of the deployment accordingly.
func (r *Resources) InspectPVCs(ctx context.Context) (util.Interval, error) {
	log := r.log
	nextInterval := maxPVCInspectorInterval

	pvcs, err := r.context.GetOwnedPVCs()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get owned PVCs")
		return 0, maskAny(err)
	}

	// Update member status from all pods found
	status, _ := r.context.GetStatus()
	for _, p := range pvcs {
		// PVC belongs to this deployment, update metric
		inspectedPVCCounter.Inc()

		// Find member status
		memberStatus, group, found := status.Members.MemberStatusByPVCName(p.GetName())
		if !found {
			log.Debug().Str("pvc", p.GetName()).Msg("no memberstatus found for PVC")
			if k8sutil.IsPersistentVolumeClaimMarkedForDeletion(&p) && len(p.GetFinalizers()) > 0 {
				// Strange, pvc belongs to us, but we have no member for it.
				// Remove all finalizers, so it can be removed.
				log.Warn().Msg("PVC belongs to this deployment, but we don't know the member. Removing all finalizers")
				kubecli := r.context.GetKubeCli()
				ignoreNotFound := false
				if err := k8sutil.RemovePVCFinalizers(log, kubecli, &p, p.GetFinalizers(), ignoreNotFound); err != nil {
					log.Debug().Err(err).Msg("Failed to update PVC (to remove all finalizers)")
					return 0, maskAny(err)
				}
			}
			continue
		}

		updateMemberStatusNeeded := false
		if k8sutil.IsPersistentVolumeClaimMarkedForDeletion(&p) {
			// Process finalizers
			if x, err := r.runPVCFinalizers(ctx, &p, group, memberStatus); err != nil {
				// Only log here, since we'll be called to try again.
				log.Warn().Err(err).Msg("Failed to run PVC finalizers")
			} else {
				nextInterval = nextInterval.ReduceTo(x)
			}
		}
		if updateMemberStatusNeeded {
			if err := status.Members.Update(memberStatus, group); err != nil {
				return 0, maskAny(err)
			}
		}
	}

	return nextInterval, nil
}
