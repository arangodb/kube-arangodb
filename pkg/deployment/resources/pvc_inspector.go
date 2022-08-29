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

package resources

import (
	"context"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	pvcv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim/v1"
)

var (
	inspectedPVCsCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_pvcs", "Number of PVC inspections per deployment", metrics.DeploymentName)
	inspectPVCsDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_pvcs_duration", "Amount of time taken by a single inspection of all PVCs for a deployment (in sec)", metrics.DeploymentName)
)

const (
	maxPVCInspectorInterval = util.Interval(time.Hour) // Maximum time between PVC inspection (if nothing else happens)
)

// InspectPVCs lists all PVCs that belong to the given deployment and updates
// the member status of the deployment accordingly.
func (r *Resources) InspectPVCs(ctx context.Context, cachedStatus inspectorInterface.Inspector) (util.Interval, error) {
	log := r.log.Str("section", "pvc")
	start := time.Now()
	nextInterval := maxPVCInspectorInterval
	deploymentName := r.context.GetAPIObject().GetName()
	defer metrics.SetDuration(inspectPVCsDurationGauges.WithLabelValues(deploymentName), start)

	// Update member status from all pods found
	status := r.context.GetStatus()
	if err := cachedStatus.PersistentVolumeClaim().V1().Iterate(func(pvc *core.PersistentVolumeClaim) error {
		// PVC belongs to this deployment, update metric
		inspectedPVCsCounters.WithLabelValues(deploymentName).Inc()

		// Find member status
		memberStatus, group, found := status.Members.MemberStatusByPVCName(pvc.GetName())
		if !found {
			log.Str("pvc", pvc.GetName()).Debug("no memberstatus found for PVC")
			if k8sutil.IsPersistentVolumeClaimMarkedForDeletion(pvc) && len(pvc.GetFinalizers()) > 0 {
				// Strange, pvc belongs to us, but we have no member for it.
				// Remove all finalizers, so it can be removed.
				log.Warn("PVC belongs to this deployment, but we don't know the member. Removing all finalizers")
				_, err := k8sutil.RemovePVCFinalizers(ctx, r.context.ACS().CurrentClusterCache(), cachedStatus.PersistentVolumeClaimsModInterface().V1(), pvc, pvc.GetFinalizers(), false)
				if err != nil {
					log.Err(err).Debug("Failed to update PVC (to remove all finalizers)")
					return errors.WithStack(err)
				}
			}
			return nil
		}

		owner := r.context.GetAPIObject().AsOwner()
		if k8sutil.UpdateOwnerRefToObjectIfNeeded(pvc.GetObjectMeta(), &owner) {
			q := patch.NewPatch(patch.ItemReplace(patch.NewPath("metadata", "ownerReferences"), pvc.ObjectMeta.OwnerReferences))
			d, err := q.Marshal()
			if err != nil {
				log.Err(err).Debug("Failed to prepare PVC patch (ownerReferences)")
				return errors.WithStack(err)
			}

			err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := cachedStatus.PersistentVolumeClaimsModInterface().V1().Patch(ctxChild, pvc.GetName(), types.JSONPatchType, d, meta.PatchOptions{})
				return err
			})

			if err != nil {
				log.Err(err).Debug("Failed to update PVC (ownerReferences)")
				return errors.WithStack(err)
			}
		}

		if k8sutil.IsPersistentVolumeClaimMarkedForDeletion(pvc) {
			// Process finalizers
			if x, err := r.runPVCFinalizers(ctx, pvc, group, memberStatus); err != nil {
				// Only log here, since we'll be called to try again.
				log.Err(err).Warn("Failed to run PVC finalizers")
			} else {
				nextInterval = nextInterval.ReduceTo(x)
			}
		}

		return nil
	}, pvcv1.FilterPersistentVolumeClaimsByLabels(k8sutil.LabelsForDeployment(deploymentName, ""))); err != nil {
		return 0, err
	}

	return nextInterval, nil
}
