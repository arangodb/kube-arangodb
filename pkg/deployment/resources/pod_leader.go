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
	"net/http"
	"sync"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

// EnsureLeader creates leader label on the pod's agency and creates service to it.
// When agency leader is not known then all agencies' pods should not have leader label, and
// consequentially service will not point to any pod.
// It works only in active fail-over mode.
func (r *Resources) EnsureLeader(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	if r.context.GetSpec().GetMode() != api.DeploymentModeActiveFailover {
		return nil
	}

	log := r.log.Str("section", "pod")

	cache, ok := r.context.GetAgencyHealth()
	if !ok {
		return nil
	}

	leaderID := cache.LeaderID()
	status, _ := r.context.GetStatus()
	noLeader := len(leaderID) == 0
	changed := false
	group := api.ServerGroupAgents
	for _, e := range status.Members.AsListInGroup(group) {
		pod, exist := cachedStatus.Pod().V1().GetSimple(e.Member.PodName)
		if !exist {
			continue
		}

		labels := pod.GetLabels()
		if noLeader || e.Member.ID != leaderID {
			// Unset a leader when:
			// - leader is unknown.
			// - leader does not belong to the current pod.

			if _, ok := labels[k8sutil.LabelKeyArangoLeader]; ok {
				delete(labels, k8sutil.LabelKeyArangoLeader)

				err := r.context.ApplyPatchOnPod(ctx, pod, patch.ItemReplace(patch.NewPath("metadata", "labels"), labels))
				if err != nil {
					log.Err(err).Error("Unable to remove leader label")
					return err
				}

				log.Warn("leader label is removed from \"%s\" member", e.Member.ID)
				changed = true
			}

			continue
		}

		// From here on it is known that there is a leader, and it should be attached to the current pod.
		if value, ok := labels[k8sutil.LabelKeyArangoLeader]; !ok {
			labels = addLabel(labels, k8sutil.LabelKeyArangoLeader, "true")
		} else if value != "true" {
			labels = addLabel(labels, k8sutil.LabelKeyArangoLeader, "true")
		} else {
			// A pod is already a leader, so nothing to change.
			continue
		}

		err := r.context.ApplyPatchOnPod(ctx, pod, patch.ItemReplace(patch.NewPath("metadata", "labels"), labels))
		if err != nil {
			log.Err(err).Error("Unable to update leader label")
			return err
		}
		log.Warn("leader label is set on \"%s\" member", e.Member.ID)
		changed = true
	}

	if changed {
		return errors.Reconcile()
	}

	if noLeader {
		// There is no leader agency so service may not exist, or it can exist with empty list of endpoints.
		return nil
	}

	leaderAgentSvcName := k8sutil.CreateAgentLeaderServiceName(r.context.GetAPIObject().GetName())
	deploymentName := r.context.GetAPIObject().GetName()

	selector := k8sutil.LabelsForLeaderMember(deploymentName, group.AsRole(), leaderID)
	if s, ok := cachedStatus.Service().V1().GetSimple(leaderAgentSvcName); ok {
		if err, adjusted := r.adjustService(ctx, s, shared.ArangoPort, selector); err == nil {
			if !adjusted {
				// The service is not changed, so single server leader can be set.
				return r.ensureSingleServerLeader(ctx, cachedStatus)
			}

			return errors.Reconcile()
		} else {
			return err
		}
	}

	s := r.createService(leaderAgentSvcName, r.context.GetNamespace(), r.context.GetAPIObject().AsOwner(), shared.ArangoPort, selector)
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := cachedStatus.ServicesModInterface().V1().Create(ctxChild, s, meta.CreateOptions{})
		return err
	})
	if err != nil {
		if !k8sutil.IsConflict(err) {
			return err
		}
	}

	// The service has been created.
	return errors.Reconcile()
}

// getSingleServerLeaderID returns id of a single server leader.
func (r *Resources) getSingleServerLeaderID(ctx context.Context) (string, error) {
	status, _ := r.context.GetStatus()
	var mutex sync.Mutex
	var leaderID string
	var anyError error

	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for _, m := range status.Members.Single {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			err := globals.GetGlobalTimeouts().ArangoD().RunWithTimeout(ctxCancel, func(ctxChild context.Context) error {
				c, err := r.context.GetMembersState().GetMemberClient(id)
				if err != nil {
					return err
				}

				if available, err := isServerAvailable(ctxChild, c); err != nil {
					return err
				} else if !available {
					return errors.New("not available")
				}

				// Other requests can be interrupted, because a leader is known already.
				cancel()
				mutex.Lock()
				leaderID = id
				mutex.Unlock()
				return nil
			})

			if err != nil {
				mutex.Lock()
				anyError = err
				mutex.Unlock()
			}
		}(m.ID)
	}
	wg.Wait()

	if len(leaderID) > 0 {
		return leaderID, nil
	}

	if anyError != nil {
		return "", errors.WithMessagef(anyError, "unable to get a leader")
	}

	return "", errors.New("unable to get a leader")
}

// setSingleServerLeadership adds or removes leadership label on a single server pod.
func (r *Resources) ensureSingleServerLeader(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	changed := false

	enabled := features.FailoverLeadership().Enabled()
	var leaderID string
	if enabled {
		var err error
		if leaderID, err = r.getSingleServerLeaderID(ctx); err != nil {
			return err
		}
	}

	status, _ := r.context.GetStatus()
	for _, m := range status.Members.Single {
		pod, exist := cachedStatus.Pod().V1().GetSimple(m.PodName)
		if !exist {
			continue
		}

		labels := pod.GetLabels()
		if enabled && m.ID == leaderID {
			if value, ok := labels[k8sutil.LabelKeyArangoLeader]; ok && value == "true" {
				// Single server is available, and it has a leader label.
				continue
			}

			labels = addLabel(labels, k8sutil.LabelKeyArangoLeader, "true")
		} else {
			if _, ok := labels[k8sutil.LabelKeyArangoLeader]; !ok {
				// Single server is not available, and it does not have a leader label.
				continue
			}

			delete(labels, k8sutil.LabelKeyArangoLeader)
		}

		err := r.context.ApplyPatchOnPod(ctx, pod, patch.ItemReplace(patch.NewPath("metadata", "labels"), labels))
		if err != nil {
			return errors.WithMessagef(err, "unable to change leader label for pod %s", m.PodName)
		}
		changed = true
	}

	if changed {
		return errors.Reconcile()
	}

	return r.ensureSingleServerLeaderServices(ctx, cachedStatus)
}

// ensureSingleServerLeaderServices adds a leadership label to deployment service and external deployment service.
func (r *Resources) ensureSingleServerLeaderServices(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	// Add a leadership label to deployment service and external deployment service.
	deploymentName := r.context.GetAPIObject().GetName()
	changed := false
	services := []string{
		k8sutil.CreateDatabaseClientServiceName(deploymentName),
		k8sutil.CreateDatabaseExternalAccessServiceName(deploymentName),
	}

	enabled := features.FailoverLeadership().Enabled()
	for _, svcName := range services {
		svc, exists := cachedStatus.Service().V1().GetSimple(svcName)
		if !exists {
			// It will be created later with a leadership label.
			continue
		}
		selector := svc.Spec.Selector
		if enabled {
			if v, ok := selector[k8sutil.LabelKeyArangoLeader]; ok && v == "true" {
				// It is already OK.
				continue
			}

			selector = addLabel(selector, k8sutil.LabelKeyArangoLeader, "true")
		} else {
			if _, ok := selector[k8sutil.LabelKeyArangoLeader]; !ok {
				// Service does not have a leader label, and it should not have.
				continue
			}

			delete(selector, k8sutil.LabelKeyArangoLeader)
		}

		parser := patch.Patch([]patch.Item{patch.ItemReplace(patch.NewPath("spec", "selector"), selector)})
		data, err := parser.Marshal()
		if err != nil {
			return errors.WithMessagef(err, "unable to marshal labels for service %s", svcName)
		}

		err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := cachedStatus.ServicesModInterface().V1().Patch(ctxChild, svcName, types.JSONPatchType, data, meta.PatchOptions{})
			return err
		})
		if err != nil {
			return errors.WithMessagef(err, "unable to patch labels for service %s", svcName)
		}
		changed = true
	}

	if changed {
		return errors.Reconcile()
	}

	return nil
}

// isServerAvailable returns true when server is available.
// In active fail-over mode one of the server should be available.
func isServerAvailable(ctx context.Context, c driver.Client) (bool, error) {
	req, err := c.Connection().NewRequest("GET", "_admin/server/availability")
	if err != nil {
		return false, errors.WithStack(err)
	}

	resp, err := c.Connection().Do(ctx, req)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if err := resp.CheckStatus(http.StatusOK, http.StatusServiceUnavailable); err != nil {
		return false, errors.WithStack(err)
	}

	return resp.StatusCode() == http.StatusOK, nil
}
