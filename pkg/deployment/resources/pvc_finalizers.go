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

	"github.com/rs/zerolog"
	"k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// runPVCFinalizers goes through the list of PVC finalizers to see if they can be removed.
func (r *Resources) runPVCFinalizers(ctx context.Context, p *v1.PersistentVolumeClaim, group api.ServerGroup, memberStatus api.MemberStatus) error {
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
		if err := k8sutil.RemovePVCFinalizers(log, kubecli, p, removalList); err != nil {
			log.Debug().Err(err).Msg("Failed to update PVC (to remove finalizers)")
			return maskAny(err)
		}
	}
	return nil
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
	// We do allow to rebuild agents
	if group == api.ServerGroupAgents && memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
		log.Debug().Msg("Rebuilding terminated agents is allowed, safe to remove member-exists finalizer")
		return nil
	}
	return maskAny(fmt.Errorf("Member still exists"))
}
