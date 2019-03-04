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

	"github.com/rs/zerolog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	podFinalizerRemovedInterval = util.Interval(time.Second / 2)  // Interval used (until new inspection) when Pod finalizers have been removed
	recheckPodFinalizerInterval = util.Interval(time.Second * 10) // Interval used when Pod finalizers need to be rechecked soon
)

// runPodFinalizers goes through the list of pod finalizers to see if they can be removed.
// Returns: Interval_till_next_inspection, error
func (r *Resources) runPodFinalizers(ctx context.Context, p *v1.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) (util.Interval, error) {
	log := r.log.With().Str("pod-name", p.GetName()).Logger()
	var removalList []string
	for _, f := range p.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerPodAgencyServing:
			log.Debug().Msg("Inspecting agency-serving finalizer")
			if err := r.inspectFinalizerPodAgencyServing(ctx, log, p, memberStatus, updateMember); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove finalizer yet")
			}
		case constants.FinalizerPodDrainDBServer:
			log.Debug().Msg("Inspecting drain dbserver finalizer")
			if err := r.inspectFinalizerPodDrainDBServer(ctx, log, p, memberStatus, updateMember); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove Pod finalizer yet")
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		kubecli := r.context.GetKubeCli()
		ignoreNotFound := false
		if err := k8sutil.RemovePodFinalizers(log, kubecli, p, removalList, ignoreNotFound); err != nil {
			log.Debug().Err(err).Msg("Failed to update pod (to remove finalizers)")
			return 0, maskAny(err)
		}
		log.Debug().Strs("finalizers", removalList).Msg("Removed finalizer(s) from Pod")
		// Let's do the next inspection quickly, since things may have changed now.
		return podFinalizerRemovedInterval, nil
	}
	// Check again at given interval
	return recheckPodFinalizerInterval, nil
}

// inspectFinalizerPodAgencyServing checks the finalizer condition for agency-serving.
// It returns nil if the finalizer can be removed.
func (r *Resources) inspectFinalizerPodAgencyServing(ctx context.Context, log zerolog.Logger, p *v1.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) error {
	if err := r.prepareAgencyPodTermination(ctx, log, p, memberStatus, func(update api.MemberStatus) error {
		if err := updateMember(update); err != nil {
			return maskAny(err)
		}
		memberStatus = update
		return nil
	}); err != nil {
		// Pod cannot be terminated yet
		return maskAny(err)
	}

	// Remaining agents are healthy, if we need to perform complete recovery
	// of the agent, also remove the PVC
	if memberStatus.Conditions.IsTrue(api.ConditionTypeAgentRecoveryNeeded) {
		pvcs := r.context.GetKubeCli().CoreV1().PersistentVolumeClaims(r.context.GetNamespace())
		if err := pvcs.Delete(memberStatus.PersistentVolumeClaimName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
			log.Warn().Err(err).Msg("Failed to delete PVC for member")
			return maskAny(err)
		}
		log.Debug().Str("pvc-name", memberStatus.PersistentVolumeClaimName).Msg("Removed PVC of member so agency can be completely replaced")
	}

	return nil
}

// inspectFinalizerPodDrainDBServer checks the finalizer condition for drain-dbserver.
// It returns nil if the finalizer can be removed.
func (r *Resources) inspectFinalizerPodDrainDBServer(ctx context.Context, log zerolog.Logger, p *v1.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) error {
	if err := r.prepareDBServerPodTermination(ctx, log, p, memberStatus, func(update api.MemberStatus) error {
		if err := updateMember(update); err != nil {
			return maskAny(err)
		}
		memberStatus = update
		return nil
	}); err != nil {
		// Pod cannot be terminated yet
		return maskAny(err)
	}

	// If this DBServer is cleaned out, we need to remove the PVC.
	if memberStatus.Conditions.IsTrue(api.ConditionTypeCleanedOut) || memberStatus.Phase == api.MemberPhaseDrain {
		pvcs := r.context.GetKubeCli().CoreV1().PersistentVolumeClaims(r.context.GetNamespace())
		if err := pvcs.Delete(memberStatus.PersistentVolumeClaimName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
			log.Warn().Err(err).Msg("Failed to delete PVC for member")
			return maskAny(err)
		}
		log.Debug().Str("pvc-name", memberStatus.PersistentVolumeClaimName).Msg("Removed PVC of member")
	}

	return nil
}
