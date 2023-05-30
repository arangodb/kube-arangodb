//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

const (
	recheckPVCFinalizerInterval = util.Interval(time.Second * 5) // Interval used when PVC finalizers need to be rechecked soon
)

// runPVCFinalizers goes through the list of PVC finalizers to see if they can be removed.
func (r *Resources) runPVCFinalizers(ctx context.Context, p *core.PersistentVolumeClaim, group api.ServerGroup,
	memberStatus api.MemberStatus) (util.Interval, error) {
	log := r.log.Str("section", "pvc").Str("pvc-name", p.GetName())
	var removalList []string
	for _, f := range p.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerPVCMemberExists:
			log.Debug("Inspecting member exists finalizer")
			if err := r.inspectFinalizerPVCMemberExists(ctx, group, memberStatus); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Err(err).Str("finalizer", f).Debug("Cannot remove finalizer yet")
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		_, err := k8sutil.RemovePVCFinalizers(ctx, r.context.ACS().CurrentClusterCache(), r.context.ACS().CurrentClusterCache().PersistentVolumeClaimsModInterface().V1(), p, removalList, false)
		if err != nil {
			log.Err(err).Debug("Failed to update PVC (to remove finalizers)")
			return 0, errors.WithStack(err)
		}
	} else {
		// Check again at given interval
		return recheckPVCFinalizerInterval, nil
	}
	return maxPVCInspectorInterval, nil
}

// inspectFinalizerPVCMemberExists checks the finalizer condition for member-exists.
// It returns nil if the finalizer can be removed.
func (r *Resources) inspectFinalizerPVCMemberExists(ctx context.Context, group api.ServerGroup,
	memberStatus api.MemberStatus) error {
	log := r.log.Str("section", "pvc")

	// Inspect member phase
	if memberStatus.Phase.IsFailed() {
		log.Debug("Member is already failed, safe to remove member-exists finalizer")
		return nil
	}

	if memberStatus.Conditions.IsTrue(api.ConditionTypeMemberVolumeUnschedulable) &&
		!memberStatus.Conditions.IsTrue(api.ConditionTypeScheduled) {
		log.Debug("Member is not scheduled and Volume is unschedulable")
		return nil
	}

	// Inspect deployment deletion state
	apiObject := r.context.GetAPIObject()
	if apiObject.GetDeletionTimestamp() != nil {
		log.Debug("Entire deployment is being deleted, safe to remove member-exists finalizer")
		return nil
	}

	// We do allow to rebuild agents & replace dbservers
	switch group {
	case api.ServerGroupAgents:
		if memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
			log.Debug("Rebuilding terminated agents is allowed, safe to remove member-exists finalizer")
			return nil
		}
	case api.ServerGroupDBServers:
		if memberStatus.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
			log.Debug("Removing cleanedout dbservers is allowed, safe to remove member-exists finalizer")
			return nil
		}
	}

	// Member still exists, let's trigger a delete of it
	if memberStatus.Pod.GetName() != "" {
		log.Info("Removing Pod of member, because PVC is being removed")
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return r.context.ACS().CurrentClusterCache().PodsModInterface().V1().Delete(ctxChild, memberStatus.Pod.GetName(), meta.DeleteOptions{})
		})
		if err != nil && !kerrors.IsNotFound(err) {
			log.Err(err).Debug("Failed to delete pod")
			return errors.WithStack(err)
		}
	}

	return errors.WithStack(errors.Newf("Member still exists"))
}
