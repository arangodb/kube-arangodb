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
	apiv1 "k8s.io/api/core/v1"
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
func (r *Resources) InspectPVCs(ctx context.Context) (util.Interval, error) {
	log := r.log
	start := time.Now()
	nextInterval := maxPVCInspectorInterval
	deploymentName := r.context.GetAPIObject().GetName()
	defer metrics.SetDuration(inspectPVCsDurationGauges.WithLabelValues(deploymentName), start)

	pvcs, err := r.context.GetOwnedPVCs()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get owned PVCs")
		return 0, maskAny(err)
	}

	// Update member status from all pods found
	status, _ := r.context.GetStatus()
	spec := r.context.GetSpec()
	for _, p := range pvcs {
		// PVC belongs to this deployment, update metric
		inspectedPVCsCounters.WithLabelValues(deploymentName).Inc()

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

		// Resize inspector
		groupSpec := spec.GetServerGroupSpec(group)

		if groupSpec.HasVolumeClaimTemplate() {
			res := groupSpec.GetVolumeClaimTemplate().Spec.Resources.Requests
			// For pvc only resources.requests is mutable
			if compareResourceList(p.Spec.Resources.Requests, res) {
				p.Spec.Resources.Requests = res
				log.Debug().Msg("volumeClaimTemplate requested resources changed - updating")
				kube := r.context.GetKubeCli()
				if _, err := kube.CoreV1().PersistentVolumeClaims(r.context.GetNamespace()).Update(&p); err != nil {
					log.Error().Err(err).Msg("Failed to update pvc")
				} else {
					r.context.CreateEvent(k8sutil.NewPVCResizedEvent(r.context.GetAPIObject(), p.Name))
				}
			}
		} else {
			if requestedSize, ok := groupSpec.Resources.Requests[apiv1.ResourceStorage]; ok {
				if volumeSize, ok := p.Spec.Resources.Requests[apiv1.ResourceStorage]; ok {
					cmp := volumeSize.Cmp(requestedSize)
					if cmp < 0 {
						// Size of the volume is smaller than the requested size
						// Update the pvc with the request size
						p.Spec.Resources.Requests[apiv1.ResourceStorage] = requestedSize

						log.Debug().Str("pvc-capacity", volumeSize.String()).Str("requested", requestedSize.String()).Msg("PVC capacity differs - updating")
						kube := r.context.GetKubeCli()
						if _, err := kube.CoreV1().PersistentVolumeClaims(r.context.GetNamespace()).Update(&p); err != nil {
							log.Error().Err(err).Msg("Failed to update pvc")
						} else {
							r.context.CreateEvent(k8sutil.NewPVCResizedEvent(r.context.GetAPIObject(), p.Name))
						}
					} else if cmp > 0 {
						log.Error().Str("server-group", group.AsRole()).Str("pvc-storage-size", volumeSize.String()).Str("requested-size", requestedSize.String()).
							Msg("Volume size should not shrink")
						r.context.CreateEvent(k8sutil.NewCannotShrinkVolumeEvent(r.context.GetAPIObject(), p.Name))
					}
				}
			}
		}

		if k8sutil.IsPersistentVolumeClaimMarkedForDeletion(&p) {
			// Process finalizers
			if x, err := r.runPVCFinalizers(ctx, &p, group, memberStatus); err != nil {
				// Only log here, since we'll be called to try again.
				log.Warn().Err(err).Msg("Failed to run PVC finalizers")
			} else {
				nextInterval = nextInterval.ReduceTo(x)
			}
		}
	}

	return nextInterval, nil
}

func compareResourceList(wanted, given apiv1.ResourceList) bool {
	for k, v := range wanted {
		if gv, ok := given[k]; !ok {
			return true
		} else if v.Cmp(gv) != 0 {
			return true
		}
	}

	for k := range given {
		if _, ok := wanted[k]; !ok {
			return true
		}
	}

	return false
}
