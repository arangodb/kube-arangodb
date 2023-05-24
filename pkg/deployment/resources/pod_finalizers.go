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
	podFinalizerRemovedInterval = util.Interval(time.Second / 2)  // Interval used (until new inspection) when Pod finalizers have been removed
	recheckPodFinalizerInterval = util.Interval(time.Second * 10) // Interval used when Pod finalizers need to be rechecked soon
	podUnreachableGracePeriod   = time.Second * 15                // Interval used when Pod finalizers need to be rechecked soon
)

// runPodFinalizers goes through the list of pod finalizers to see if they can be removed.
// Returns: Interval_till_next_inspection, error
func (r *Resources) runPodFinalizers(ctx context.Context, p *core.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) (util.Interval, error) {
	log := r.log.Str("section", "pod").Str("pod-name", p.GetName())
	var removalList []string

	// When the main container is terminated, then the whole pod should be terminated,
	// so sidecar core containers' names should not be checked here.
	// If Member is not reachable finalizers should be also removed
	isServerContainerDead := !k8sutil.IsPodServerContainerRunning(p) || memberStatus.Conditions.Check(api.ConditionTypeReachable).Exists().IsFalse().LastTransition(podUnreachableGracePeriod).Evaluate()

	for _, f := range p.ObjectMeta.GetFinalizers() {
		switch f {
		case constants.FinalizerPodAgencyServing:
			log.Debug("Inspecting agency-serving finalizer")
			if isServerContainerDead {
				log.Debug("Server Container is dead, removing finalizer")
				removalList = append(removalList, f)
				break
			}
			if err := r.inspectFinalizerPodAgencyServing(ctx, p, memberStatus, updateMember); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Err(err).Str("finalizer", f).Debug("Cannot remove finalizer yet")
			}
		case constants.FinalizerPodDrainDBServer:
			log.Debug("Inspecting drain dbserver finalizer")
			if isServerContainerDead {
				log.Debug("Server Container is dead, removing finalizer")
				removalList = append(removalList, f)
				break
			}
			if err := r.inspectFinalizerPodDrainDBServer(ctx, p, memberStatus, updateMember); err == nil {
				removalList = append(removalList, f)
			} else {
				log.Err(err).Str("finalizer", f).Debug("Cannot remove Pod finalizer yet")
			}
		case constants.FinalizerPodGracefulShutdown:
			// We are in graceful shutdown, only one way to remove it is when container is already dead
			if isServerContainerDead {
				log.Debug("Server Container is dead, removing finalizer")
				removalList = append(removalList, f)
			}
		case constants.FinalizerDelayPodTermination:
			if isServerContainerDead {
				log.Debug("Server Container is dead, removing finalizer")
				removalList = append(removalList, f)
				break
			}

			s := r.context.GetStatus()
			_, group, ok := s.Members.ElementByID(memberStatus.ID)
			if !ok {
				continue
			}
			log.Str("finalizer", f).Error("Delay finalizer")

			groupSpec := r.context.GetSpec().GetServerGroupSpec(group)
			if t := p.ObjectMeta.DeletionTimestamp; t != nil {
				d := time.Duration(groupSpec.GetShutdownDelay(group)) * time.Second
				gr := time.Duration(util.TypeOrDefault[int64](p.ObjectMeta.GetDeletionGracePeriodSeconds(), 0)) * time.Second
				e := t.Time.Add(-1 * gr).Sub(time.Now().Add(-1 * d))
				log.Str("finalizer", f).Str("left", e.String()).Error("Delay finalizer status")
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
		if _, err := k8sutil.RemovePodFinalizers(ctx, r.context.ACS().CurrentClusterCache(), r.context.ACS().CurrentClusterCache().PodsModInterface().V1(), p, removalList, false); err != nil {
			log.Err(err).Debug("Failed to update pod (to remove finalizers)")
			return 0, errors.WithStack(err)
		}
		log.Strs("finalizers", removalList...).Debug("Removed finalizer(s) from Pod")
		// Let's do the next inspection quickly, since things may have changed now.
		return podFinalizerRemovedInterval, nil
	}
	// Check again at given interval
	return recheckPodFinalizerInterval, nil
}

// inspectFinalizerPodAgencyServing checks the finalizer condition for agency-serving.
// It returns nil if the finalizer can be removed.
func (r *Resources) inspectFinalizerPodAgencyServing(ctx context.Context, p *core.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) error {
	log := r.log.Str("section", "agency")
	if err := r.prepareAgencyPodTermination(p, memberStatus, func(update api.MemberStatus) error {
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
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return r.context.ACS().CurrentClusterCache().PersistentVolumeClaimsModInterface().V1().Delete(ctxChild, memberStatus.PersistentVolumeClaim.GetName(), meta.DeleteOptions{})
		})
		if err != nil && !kerrors.IsNotFound(err) {
			log.Err(err).Warn("Failed to delete PVC for member")
			return errors.WithStack(err)
		}
		log.Str("pvc-name", memberStatus.PersistentVolumeClaim.GetName()).Debug("Removed PVC of member so agency can be completely replaced")
	}

	return nil
}

// inspectFinalizerPodDrainDBServer checks the finalizer condition for drain-dbserver.
// It returns nil if the finalizer can be removed.
func (r *Resources) inspectFinalizerPodDrainDBServer(ctx context.Context, p *core.Pod, memberStatus api.MemberStatus, updateMember func(api.MemberStatus) error) error {
	log := r.log.Str("section", "pod")
	if err := r.prepareDBServerPodTermination(ctx, p, memberStatus, func(update api.MemberStatus) error {
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
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return r.context.ACS().CurrentClusterCache().PersistentVolumeClaimsModInterface().V1().Delete(ctxChild, memberStatus.PersistentVolumeClaim.GetName(), meta.DeleteOptions{})
		})
		if err != nil && !kerrors.IsNotFound(err) {
			log.Err(err).Warn("Failed to delete PVC for member")
			return errors.WithStack(err)
		}
		log.Str("pvc-name", memberStatus.PersistentVolumeClaim.GetName()).Debug("Removed PVC of member")
	}

	return nil
}
