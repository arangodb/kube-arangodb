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

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	recheckPVCFinalizerInterval = util.Interval(time.Second * 5) // Interval used when PVC finalizers need to be rechecked soon
)

// runPVCFinalizers goes through the list of PVC finalizers to see if they can be removed.
func (r *Resources) runPVCFinalizers(ctx context.Context, p *v1.PersistentVolumeClaim, group api.ServerGroup,
	memberStatus api.MemberStatus) (util.Interval, error) {
	log := r.log.With().Str("pvc-name", p.GetName()).Logger()
	var removalList []string
	for _, f := range p.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerPVCMemberExists:
			log.Debug().Msg("Inspecting member exists finalizer")
			if err := r.inspectFinalizerPVCMemberExists(ctx, log, group, memberStatus); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove finalizer yet")
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		_, err := k8sutil.RemovePVCFinalizers(ctx, r.context.GetCachedStatus(), log, r.context.PersistentVolumeClaimsModInterface(), p, removalList, false)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to update PVC (to remove finalizers)")
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
func (r *Resources) inspectFinalizerPVCMemberExists(ctx context.Context, log zerolog.Logger, group api.ServerGroup,
	memberStatus api.MemberStatus) error {
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
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return r.context.PodsModInterface().Delete(ctxChild, memberStatus.PodName, meta.DeleteOptions{})
		})
		if err != nil && !k8sutil.IsNotFound(err) {
			log.Debug().Err(err).Msg("Failed to delete pod")
			return errors.WithStack(err)
		}
	}

	return errors.WithStack(errors.Newf("Member still exists"))
}
