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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	recheckPVCFinalizerInterval = util.Interval(time.Second * 5) // Interval used when PVC finalizers need to be rechecked soon
)

// runPVCFinalizers goes through the list of PVC finalizers to see if they can be removed.
func (r *Resources) runPVCFinalizers(ctx context.Context, p *v1.PersistentVolumeClaim, group api.ServerGroup, memberStatus api.MemberStatus) (util.Interval, error) {
	log := r.log.With().Str("pvc-name", p.GetName()).Logger()
	var removalList []string
	for _, f := range p.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerPVCMemberExists:
			log.Debug().Msg("Inspecting member exists finalizer")
			if err := r.inspectFinalizerPVCMemberExists(ctx, log, p, group, memberStatus); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove finalizer yet")
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
		}
	} else {
		// Check again at given interval
		return recheckPVCFinalizerInterval, nil
	}
	return maxPVCInspectorInterval, nil
}

// inspectFinalizerPVCMemberExists checks the finalizer condition for member-exists.
// It returns nil if the finalizer can be removed.
func (r *Resources) inspectFinalizerPVCMemberExists(ctx context.Context, log zerolog.Logger, p *v1.PersistentVolumeClaim, group api.ServerGroup, memberStatus api.MemberStatus) error {
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

	// Member still exists, let's trigger a delete of it
	if memberStatus.PodName != "" {
		log.Info().Msg("Removing Pod of member, because PVC is being removed")
		pods := r.context.GetKubeCli().CoreV1().Pods(apiObject.GetNamespace())
		if err := pods.Delete(memberStatus.PodName, &metav1.DeleteOptions{}); err != nil && !k8sutil.IsNotFound(err) {
			log.Debug().Err(err).Msg("Failed to delete pod")
			return maskAny(err)
		}
	}

	return maskAny(fmt.Errorf("Member still exists"))
}
