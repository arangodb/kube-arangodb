//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package resources

import (
	"context"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
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

	isServerContainerDead := !k8sutil.IsPodServerContainerRunning(p)

	for _, f := range p.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerPodAgencyServing:
			log.Debug().Msg("Inspecting agency-serving finalizer")
			if isServerContainerDead {
				log.Debug().Msg("Server Container is dead, removing finalizer")
				removalList = append(removalList, f)
				break
			}
			if err := r.inspectFinalizerPodAgencyServing(ctx, log, p, memberStatus, updateMember); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove finalizer yet")
			}
		case constants.FinalizerPodDrainDBServer:
			log.Debug().Msg("Inspecting drain dbserver finalizer")
			if isServerContainerDead {
				log.Debug().Msg("Server Container is dead, removing finalizer")
				removalList = append(removalList, f)
				break
			}
			if err := r.inspectFinalizerPodDrainDBServer(ctx, log, p, memberStatus, updateMember); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove Pod finalizer yet")
			}
		case constants.FinalizerPodGracefulShutdown:
			// We are in graceful shutdown, only one way to remove it is when container is already dead
			if isServerContainerDead {
				log.Debug().Msg("Server Container is dead, removing finalizer")
				removalList = append(removalList, f)
			}
		case constants.FinalizerDelayPodTermination:
			if isServerContainerDead {
				log.Debug().Msg("Server Container is dead, removing finalizer")
				removalList = append(removalList, f)
				break
			}

			s, _ := r.context.GetStatus()
			_, group, ok := s.Members.ElementByID(memberStatus.ID)
			if !ok {
				continue
			}
			log.Error().Str("finalizer", f).Msg("Delay finalizer")

			groupSpec := r.context.GetSpec().GetServerGroupSpec(group)
			if t := p.ObjectMeta.DeletionTimestamp; t != nil {
				d := time.Duration(groupSpec.GetShutdownDelay(group)) * time.Second
				gr := time.Duration(util.Int64OrDefault(p.ObjectMeta.GetDeletionGracePeriodSeconds(), 0)) * time.Second
				e := t.Time.Add(-1 * gr).Sub(time.Now().Add(-1 * d))
				log.Error().Str("finalizer", f).Str("left", e.String()).Msg("Delay finalizer status")
				if e < 0 || d == 0 {
					removalList = append(removalList, f)
				}
			} else {
				continue
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		if err := k8sutil.RemovePodFinalizers(ctx, r.context.GetCachedStatus(), log, r.context.PodsModInterface(), p, removalList, false); err != nil {
			log.Debug().Err(err).Msg("Failed to update pod (to remove finalizers)")
			return 0, errors.WithStack(err)
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
			return errors.WithStack(err)
		}
		memberStatus = update
		return nil
	}); err != nil {
		// Pod cannot be terminated yet
		return errors.WithStack(err)
	}

	// Remaining agents are healthy, if we need to perform complete recovery
	// of the agent, also remove the PVC
	if memberStatus.Conditions.IsTrue(api.ConditionTypeAgentRecoveryNeeded) {
		err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return r.context.PersistentVolumeClaimsModInterface().Delete(ctxChild, memberStatus.PersistentVolumeClaimName, metav1.DeleteOptions{})
		})
		if err != nil && !k8sutil.IsNotFound(err) {
			log.Warn().Err(err).Msg("Failed to delete PVC for member")
			return errors.WithStack(err)
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
			return errors.WithStack(err)
		}
		memberStatus = update
		return nil
	}); err != nil {
		// Pod cannot be terminated yet
		return errors.WithStack(err)
	}

	// If this DBServer is cleaned out, we need to remove the PVC.
	if memberStatus.Conditions.IsTrue(api.ConditionTypeCleanedOut) || memberStatus.Phase == api.MemberPhaseDrain {
		err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return r.context.PersistentVolumeClaimsModInterface().Delete(ctxChild, memberStatus.PersistentVolumeClaimName, metav1.DeleteOptions{})
		})
		if err != nil && !k8sutil.IsNotFound(err) {
			log.Warn().Err(err).Msg("Failed to delete PVC for member")
			return errors.WithStack(err)
		}
		log.Debug().Str("pvc-name", memberStatus.PersistentVolumeClaimName).Msg("Removed PVC of member")
	}

	return nil
}
