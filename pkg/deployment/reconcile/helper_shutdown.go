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

package reconcile

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"

	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// getShutdownHelper returns an action to shut down a pod according to the settings.
// Returns true when member status exists.
// There are 3 possibilities to shut down the pod: immediately, gracefully, standard kubernetes delete API.
// When pod does not exist then success action (which always successes) is returned.
func getShutdownHelper(a *api.Action, actionCtx ActionContext, log zerolog.Logger) (ActionCore, api.MemberStatus, bool) {
	m, ok := actionCtx.GetMemberStatusByID(a.MemberID)
	if !ok {
		log.Warn().Str("pod-name", m.PodName).Msg("member is already gone")

		return nil, api.MemberStatus{}, false
	}

	pod, ok := actionCtx.GetCachedStatus().Pod(m.PodName)
	if !ok {
		log.Warn().Str("pod-name", m.PodName).Msg("pod is already gone")
		// Pod does not exist, so create success action to finish it immediately.
		return NewActionSuccess(), m, true
	}

	if _, ok := pod.GetAnnotations()[deployment.ArangoDeploymentPodDeleteNow]; ok {
		// The pod contains annotation, so pod must be deleted immediately.
		return shutdownNow{action: a, actionCtx: actionCtx, log: log, memberStatus: m}, m, true
	}

	if features.GracefulShutdown().Enabled() {
		return shutdownHelperAPI{action: a, actionCtx: actionCtx, log: log, memberStatus: m}, m, true
	}

	serverGroup := actionCtx.GetSpec().GetServerGroupSpec(a.Group)

	switch serverGroup.ShutdownMethod.Get() {
	case api.ServerGroupShutdownMethodDelete:
		return shutdownHelperDelete{action: a, actionCtx: actionCtx, log: log, memberStatus: m}, m, true
	default:
		return shutdownHelperAPI{action: a, actionCtx: actionCtx, log: log, memberStatus: m}, m, true
	}
}

type shutdownHelperAPI struct {
	log          zerolog.Logger
	action       *api.Action
	actionCtx    ActionContext
	memberStatus api.MemberStatus
}

func (s shutdownHelperAPI) Start(ctx context.Context) (bool, error) {
	log := s.log

	log.Info().Msgf("Using API to shutdown member")

	group := s.action.Group
	podName := s.memberStatus.PodName
	if podName == "" {
		log.Warn().Msgf("Pod is empty")
		return true, nil
	}
	// Remove finalizers, so Kubernetes will quickly terminate the pod
	if !features.GracefulShutdown().Enabled() {
		if err := s.actionCtx.RemovePodFinalizers(ctx, podName); err != nil {
			return false, errors.WithStack(err)
		}
	}

	if group.IsArangod() {
		// Invoke shutdown endpoint
		ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		c, err := s.actionCtx.GetServerClient(ctxChild, group, s.action.MemberID)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create member client")
			return false, errors.WithStack(err)
		}
		removeFromCluster := false
		log.Debug().Bool("removeFromCluster", removeFromCluster).Msg("Shutting down member")
		ctxChild, cancel = context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()
		if err := c.ShutdownV2(ctxChild, removeFromCluster, true); err != nil {
			// Shutdown failed. Let's check if we're already done
			if ready, _, err := s.CheckProgress(ctxChild); err == nil && ready {
				// We're done
				return true, nil
			}
			log.Debug().Err(err).Msg("Failed to shutdown member")
			return false, errors.WithStack(err)
		}
	} else if group.IsArangosync() {
		// Terminate pod
		if err := s.actionCtx.DeletePod(ctx, podName, meta.DeleteOptions{}); err != nil {
			return false, errors.WithStack(err)
		}
	}

	return false, nil
}

// CheckProgress returns true when pod is terminated.
func (s shutdownHelperAPI) CheckProgress(_ context.Context) (bool, bool, error) {
	terminated := s.memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated)
	return terminated, false, nil
}

type shutdownHelperDelete struct {
	log          zerolog.Logger
	action       *api.Action
	actionCtx    ActionContext
	memberStatus api.MemberStatus
}

func (s shutdownHelperDelete) Start(ctx context.Context) (bool, error) {
	log := s.log

	log.Info().Msgf("Using Pod Delete to shutdown member")

	podName := s.memberStatus.PodName
	if podName == "" {
		log.Warn().Msgf("Pod is empty")
		return true, nil
	}

	// Terminate pod
	if err := s.actionCtx.DeletePod(ctx, podName, meta.DeleteOptions{}); err != nil {
		if !k8sutil.IsNotFound(err) {
			return false, errors.WithStack(err)
		}
	}

	return false, nil
}

func (s shutdownHelperDelete) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Check that pod is removed
	log := s.log
	if !s.memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
		// Pod is not yet terminated
		log.Warn().Msgf("Pod not yet terminated")
		return false, false, nil
	}

	podName := s.memberStatus.PodName
	if podName != "" {
		if _, err := s.actionCtx.GetPod(ctx, podName); err == nil {
			log.Warn().Msgf("Pod still exists")
			return false, false, nil
		} else if !k8sutil.IsNotFound(err) {
			log.Error().Err(err).Msg("Unable to get pod")
			return false, false, nil
		}
	}

	return true, false, nil
}

type shutdownNow struct {
	action       *api.Action
	actionCtx    ActionContext
	memberStatus api.MemberStatus
	log          zerolog.Logger
}

// Start starts removing pod forcefully.
func (s shutdownNow) Start(ctx context.Context) (bool, error) {
	// Check progress is used here because removing pod can start gracefully,
	// and then it can be changed to force shutdown.
	s.log.Info().Msg("Using shutdown now method")
	ready, _, err := s.CheckProgress(ctx)
	return ready, err
}

// CheckProgress starts removing pod forcefully and checks if has it been removed.
func (s shutdownNow) CheckProgress(ctx context.Context) (bool, bool, error) {
	podName := s.memberStatus.PodName
	pod, err := s.actionCtx.GetPod(ctx, podName)
	if err != nil {
		if k8sutil.IsNotFound(err) {
			s.log.Info().Msg("Using shutdown now method completed because pod is gone")
			return true, false, nil
		}

		return false, false, errors.Wrapf(err, "failed to get pod")
	}

	if s.memberStatus.PodUID != pod.GetUID() {
		s.log.Info().Msg("Using shutdown now method completed because it is already rotated")
		// The new pod has been started already.
		return true, false, nil
	}

	// Remove finalizers forcefully.
	if err := s.actionCtx.RemovePodFinalizers(ctx, podName); err != nil {
		return false, false, errors.WithStack(err)
	}

	// Terminate pod.
	options := meta.DeleteOptions{
		// Leave one second to clean a PVC.
		GracePeriodSeconds: util.NewInt64(1),
	}
	if err := s.actionCtx.DeletePod(ctx, podName, options); err != nil {
		if !k8sutil.IsNotFound(err) {
			return false, false, errors.WithStack(err)
		}
	}

	s.log.Info().Msgf("Using shutdown now method completed")
	return true, false, nil
}
