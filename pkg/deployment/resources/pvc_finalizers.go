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
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	recheckPVCFinalizerInterval = time.Second * 10 // Interval used when PVC finalizers need to be rechecked soon
)

// runPVCFinalizers goes through the list of PVC finalizers to see if they can be removed.
// Returns: Interval_till_next_inspection, error
func (r *Resources) runPVCFinalizers(ctx context.Context, p *v1.PersistentVolumeClaim, group api.ServerGroup, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) (time.Duration, error) {
	log := r.log.With().Str("pvc-name", p.GetName()).Logger()
	var removalList []string
	for _, f := range p.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerPVCMemberExists:
			log.Debug().Msg("Inspecting member exists finalizer")
			if err := r.inspectFinalizerPVCMemberExists(ctx, log, p, group, memberStatus, updateMember); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove PVC finalizer yet")
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		kubecli := r.context.GetKubeCli()
		ignoreNotFound := false
		if err := k8sutil.RemovePVCFinalizers(log, kubecli, p, removalList, ignoreNotFound); err != nil {
			log.Debug().Err(err).Msg("Failed to update PVC (to remove finalizers)")
			return 0, maskAny(err)
		} else {
			log.Debug().Strs("finalizers", removalList).Msg("Removed finalizer(s) from PVC")
		}
	} else {
		// Check again at given interval
		return recheckPVCFinalizerInterval, nil
	}
	return maxPVCInspectorInterval, nil
}

// inspectFinalizerPVCMemberExists checks the finalizer condition for member-exists.
// It returns nil if the finalizer can be removed.
func (r *Resources) inspectFinalizerPVCMemberExists(ctx context.Context, log zerolog.Logger, p *v1.PersistentVolumeClaim, group api.ServerGroup, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) error {
	// Inspect member phase
	if memberStatus.Phase.IsFailed() {
		log.Debug().Msg("Member is already failed, safe to remove member-exists finalizer")
		return nil
	}
	// Inspect deployment deletion state
	apiObject := r.context.GetAPIObject()
	if apiObject.GetDeletionTimestamp() != nil {
		log.Debug().Msg("Entire deployment is being deleted, safe to remove member-exists finalizer")
		return nil
	}

	// We do allow to rebuild agents & replace dbservers
	switch group {
	case api.ServerGroupAgents:
		if memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
			log.Debug().Msg("Rebuilding terminated agents is allowed, safe to remove member-exists finalizer")
			return nil
		}
	case api.ServerGroupDBServers:
		if memberStatus.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
			log.Debug().Msg("Removing cleanedout dbservers is allowed, safe to remove member-exists finalizer")
			return nil
		}
	}

	// Member still exists, let's trigger a delete of it, if we're allowed to do so
	if memberStatus.PodName != "" {
		pods := r.context.GetKubeCli().CoreV1().Pods(apiObject.GetNamespace())
		log.Info().Msg("Checking in Pod of member can be removed, because PVC is being removed")
		if pod, err := pods.Get(memberStatus.PodName, metav1.GetOptions{}); err != nil && !k8sutil.IsNotFound(err) {
			log.Debug().Err(err).Msg("Failed to get pod for PVC")
			return maskAny(err)
		} else if err == nil {
			// We've got the pod, check & prepare its termination
			if err := r.preparePodTermination(ctx, log, pod, group, memberStatus, updateMember); err != nil {
				log.Debug().Err(err).Msg("Not allowed to remove pod yet")
				return maskAny(err)
			}
		}

		log.Info().Msg("Removing Pod of member, because PVC is being removed")
		if err := pods.Delete(memberStatus.PodName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
			log.Debug().Err(err).Msg("Failed to delete pod")
			return maskAny(err)
		}
	}

	return maskAny(fmt.Errorf("Member still exists"))
}
