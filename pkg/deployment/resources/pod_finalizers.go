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
	"net/http"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
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

	if !k8sutil.IsPodServerContainerRunning(p) {
		// When the container is dead then remove finalizers.
		for _, f := range p.ObjectMeta.GetFinalizers() {
			switch f {
			case constants.FinalizerPodAgencyServing, constants.FinalizerPodDrainDBServer,
				constants.FinalizerGracefulShutdown, constants.FinalizerDelayPodTermination:
				removalList = append(removalList, f)
				log.Debug().Str("finalizer", f).Msg("Server Container is dead, removing finalizer")
			}
		}
	} else {
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
			case constants.FinalizerGracefulShutdown:
				if err := r.inspectFinalizerGracefulShutdown(ctx, log); err == nil {
					removalList = append(removalList, f)
				} else {
					log.Debug().Err(err).Str("finalizer", f).Msg("Cannot remove Pod finalizer yet")
				}
			case constants.FinalizerDelayPodTermination:
				s, _ := r.context.GetStatus()
				_, group, ok := s.Members.ElementByID(memberStatus.ID)
				if !ok {
					continue
				}
				log.Error().Str("finalizer", f).Msg("Delay finalizer")

				groupSpec := r.context.GetSpec().GetServerGroupSpec(group)
				d := time.Duration(util.IntOrDefault(groupSpec.ShutdownDelay, 0)) * time.Second
				if t := p.ObjectMeta.DeletionTimestamp; t != nil {
					e := p.ObjectMeta.DeletionTimestamp.Time.Sub(time.Now().Add(d))
					log.Error().Str("finalizer", f).Dur("left", e).Msg("Delay finalizer status")
					if e < 0 {
						removalList = append(removalList, f)
					}
				} else {
					continue
				}
			}
		}
	}
	// Remove finalizers (if needed)
	if len(removalList) > 0 {
		kubecli := r.context.GetKubeCli()
		if err := k8sutil.RemovePodFinalizers(ctx, log, kubecli, p, removalList, false); err != nil {
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
		pvcs := r.context.GetKubeCli().CoreV1().PersistentVolumeClaims(r.context.GetNamespace())
		err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return pvcs.Delete(ctxChild, memberStatus.PersistentVolumeClaimName, metav1.DeleteOptions{})
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
		pvcs := r.context.GetKubeCli().CoreV1().PersistentVolumeClaims(r.context.GetNamespace())
		err := k8sutil.RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return pvcs.Delete(ctxChild, memberStatus.PersistentVolumeClaimName, metav1.DeleteOptions{})
		})
		if err != nil && !k8sutil.IsNotFound(err) {
			log.Warn().Err(err).Msg("Failed to delete PVC for member")
			return errors.WithStack(err)
		}
		log.Debug().Str("pvc-name", memberStatus.PersistentVolumeClaimName).Msg("Removed PVC of member")
	}

	return nil
}

// inspectFinalizerGracefulShutdown sends graceful shutdown request if it was not sent beforehand.
func (r *Resources) inspectFinalizerGracefulShutdown(ctx context.Context, log zerolog.Logger) error {
	var c driver.Client
	c, err := r.context.GetDatabaseClient(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get database client")
		return errors.WithStack(err)
	}

	if shutdownInfo, err := arangod.GetShutdownInfo(ctx, c); err != nil {
		if !driver.IsArangoErrorWithCode(err, http.StatusMethodNotAllowed) {
			log.Debug().Err(err).Msg("Failed to check shutdown info")
			return errors.WithStack(err)
		}
	} else if shutdownInfo.SoftShutdownOngoing {
		return nil
	}

	log.Debug().
		Bool("removeFromCluster", false).
		Bool("gracefulShutdown", true).
		Msg("Shutting down member")

	ctxChild, cancel := context.WithTimeout(ctx, arangod.GetRequestTimeout())
	defer cancel()
	if err := c.ShutdownV2(ctxChild, false, true); err != nil {
		log.Debug().
			Err(err).
			Bool("removeFromCluster", false).
			Bool("gracefulShutdown", true).
			Msg("Failed to shutdown member")
		return errors.WithStack(err)
	}

	return nil
}
