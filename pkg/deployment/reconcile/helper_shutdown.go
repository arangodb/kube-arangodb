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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
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
func getShutdownHelper(a actionImpl) (ActionCore, api.MemberStatus, bool) {
	m, ok := a.actionCtx.GetMemberStatusByID(a.action.MemberID)
	if !ok {
		a.log.Str("pod-name", m.PodName).Warn("member is already gone")

		return nil, api.MemberStatus{}, false
	}

	cache, ok := a.actionCtx.ACS().ClusterCache(m.ClusterID)
	if !ok {
		a.log.Str("pod-name", m.PodName).Warn("Cluster is not ready")

		return nil, api.MemberStatus{}, false
	}

	if ifPodUIDMismatch(m, a.action, cache) {
		a.log.Error("Member UID is changed")
		return NewActionSuccess(), m, true
	}

	pod, ok := cache.Pod().V1().GetSimple(m.PodName)
	if !ok {
		a.log.Str("pod-name", m.PodName).Warn("pod is already gone")
		// Pod does not exist, so create success action to finish it immediately.
		return NewActionSuccess(), m, true
	}

	if _, ok := pod.GetAnnotations()[deployment.ArangoDeploymentPodDeleteNow]; ok {
		// The pod contains annotation, so pod must be deleted immediately.
		return shutdownNow{actionImpl: a, memberStatus: m}, m, true
	}

	if features.GracefulShutdown().Enabled() {
		return shutdownHelperAPI{actionImpl: a, memberStatus: m}, m, true
	}

	serverGroup := a.actionCtx.GetSpec().GetServerGroupSpec(a.action.Group)

	switch serverGroup.ShutdownMethod.Get() {
	case api.ServerGroupShutdownMethodDelete:
		return shutdownHelperDelete{actionImpl: a, memberStatus: m}, m, true
	default:
		return shutdownHelperAPI{actionImpl: a, memberStatus: m}, m, true
	}
}

type shutdownHelperAPI struct {
	actionImpl
	memberStatus api.MemberStatus
}

func (s shutdownHelperAPI) Start(ctx context.Context) (bool, error) {
	s.log.Info("Using API to shutdown member")

	group := s.action.Group
	podName := s.memberStatus.PodName
	if podName == "" {
		s.log.Warn("Pod is empty")
		return true, nil
	}

	cache, ok := s.actionCtx.ACS().ClusterCache(s.memberStatus.ClusterID)
	if !ok {
		return true, errors.Newf("Cluster is not ready")
	}

	// Remove finalizers, so Kubernetes will quickly terminate the pod
	if !features.GracefulShutdown().Enabled() {
		pod, ok := cache.Pod().V1().GetSimple(podName)
		if ok && len(pod.Finalizers) > 0 {
			pod.Finalizers = nil

			ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
			defer cancel()

			if _, err := cache.Client().Kubernetes().CoreV1().Pods(cache.Namespace()).Update(ctxChild, pod, meta.UpdateOptions{}); err != nil {
				return false, err
			}
		}
	}

	if group.IsArangod() {
		// Invoke shutdown endpoint
		ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
		defer cancel()
		c, err := s.actionCtx.GetServerClient(ctxChild, group, s.action.MemberID)
		if err != nil {
			s.log.Err(err).Debug("Failed to create member client")
			return false, errors.WithStack(err)
		}
		removeFromCluster := false
		s.log.Bool("removeFromCluster", removeFromCluster).Debug("Shutting down member")
		ctxChild, cancel = context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()
		if err := c.ShutdownV2(ctxChild, removeFromCluster, true); err != nil {
			// Shutdown failed. Let's check if we're already done
			if ready, _, err := s.CheckProgress(ctxChild); err == nil && ready {
				// We're done
				return true, nil
			}
			s.log.Err(err).Debug("Failed to shutdown member")
			return false, errors.WithStack(err)
		}
	} else if group.IsArangosync() {
		// Terminate pod
		if err := cache.Client().Kubernetes().CoreV1().Pods(cache.Namespace()).Delete(ctx, podName, meta.DeleteOptions{}); err != nil {
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
	actionImpl
	memberStatus api.MemberStatus
}

func (s shutdownHelperDelete) Start(ctx context.Context) (bool, error) {
	s.log.Info("Using Pod Delete to shutdown member")

	podName := s.memberStatus.PodName
	if podName == "" {
		s.log.Warn("Pod is empty")
		return true, nil
	}

	cache, ok := s.actionCtx.ACS().ClusterCache(s.memberStatus.ClusterID)
	if !ok {
		return true, errors.Newf("Cluster is not ready")
	}

	// Terminate pod
	if err := cache.Client().Kubernetes().CoreV1().Pods(cache.Namespace()).Delete(ctx, podName, meta.DeleteOptions{}); err != nil {
		if !k8sutil.IsNotFound(err) {
			return false, errors.WithStack(err)
		}
	}

	return false, nil
}

func (s shutdownHelperDelete) CheckProgress(ctx context.Context) (bool, bool, error) {
	// Check that pod is removed
	if !s.memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
		// Pod is not yet terminated
		s.log.Warn("Pod not yet terminated")
		return false, false, nil
	}

	cache, ok := s.actionCtx.ACS().ClusterCache(s.memberStatus.ClusterID)
	if !ok {
		s.log.Warn("Cluster is not ready")
		return false, false, nil
	}

	podName := s.memberStatus.PodName
	if podName != "" {
		if _, ok := cache.Pod().V1().GetSimple(podName); ok {
			s.log.Warn("Pod still exists")
			return false, false, nil
		}
	}

	return true, false, nil
}

type shutdownNow struct {
	actionImpl
	memberStatus api.MemberStatus
}

// Start starts removing pod forcefully.
func (s shutdownNow) Start(ctx context.Context) (bool, error) {
	// Check progress is used here because removing pod can start gracefully,
	// and then it can be changed to force shutdown.
	s.log.Info("Using shutdown now method")
	ready, _, err := s.CheckProgress(ctx)
	return ready, err
}

// CheckProgress starts removing pod forcefully and checks if has it been removed.
func (s shutdownNow) CheckProgress(ctx context.Context) (bool, bool, error) {
	podName := s.memberStatus.PodName

	cache, ok := s.actionCtx.ACS().ClusterCache(s.memberStatus.ClusterID)
	if !ok {
		s.log.Warn("Cluster is not ready")
		return false, false, nil
	}

	pod, ok := cache.Pod().V1().GetSimple(podName)
	if !ok {
		s.log.Info("Using shutdown now method completed because pod is gone")
		return true, false, nil
	}

	if s.memberStatus.PodUID != pod.GetUID() {
		s.log.Info("Using shutdown now method completed because it is already rotated")
		// The new pod has been started already.
		return true, false, nil
	}

	// Remove finalizers forcefully.
	if len(pod.Finalizers) > 0 {
		pod.Finalizers = nil

		ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
		defer cancel()

		if _, err := cache.Client().Kubernetes().CoreV1().Pods(cache.Namespace()).Update(ctxChild, pod, meta.UpdateOptions{}); err != nil {
			return false, false, err
		}
	}

	// Terminate pod.
	options := meta.DeleteOptions{
		// Leave one second to clean a PVC.
		GracePeriodSeconds: util.NewInt64(1),
	}
	if err := cache.Client().Kubernetes().CoreV1().Pods(cache.Namespace()).Delete(ctx, podName, options); err != nil {
		if !k8sutil.IsNotFound(err) {
			return false, false, errors.WithStack(err)
		}
	}

	s.log.Info("Using shutdown now method completed")
	return true, false, nil
}
