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
	"strings"
	"sync"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
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
	status := r.context.GetStatus()
	noLeader := len(leaderID) == 0
	changed := false
	group := api.ServerGroupAgents
	for _, e := range status.Members.AsListInGroup(group) {
		pod, exist := cachedStatus.Pod().V1().GetSimple(e.Member.Pod.GetName())
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

	ports := []core.ServicePort{CreateServerServicePort()}
	selector := k8sutil.LabelsForLeaderMember(deploymentName, group.AsRole(), leaderID)

	if s, ok := cachedStatus.Service().V1().GetSimple(leaderAgentSvcName); ok {
		if c, err := patcher.ServicePatcher(ctx, cachedStatus.ServicesModInterface().V1(), s, meta.PatchOptions{}, patcher.PatchServiceSelector(selector), patcher.PatchServicePorts(ports)); err != nil {
			return err
		} else {
			if !c {
				return r.ensureSingleServerLeader(ctx, cachedStatus)
			}

			return errors.Reconcile()
		}
	}

	s := r.createService(leaderAgentSvcName, r.context.GetNamespace(), "", core.ServiceTypeClusterIP, r.context.GetAPIObject().AsOwner(), ports, selector)
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := cachedStatus.ServicesModInterface().V1().Create(ctxChild, s, meta.CreateOptions{})
		return err
	})
	if err != nil {
		if !kerrors.IsConflict(err) {
			return err
		}
	}

	// The service has been created.
	return errors.Reconcile()
}

// getSingleServerLeaderID returns ids of a single server leaders.
func (r *Resources) getSingleServerLeaderID(ctx context.Context) ([]string, error) {
	status := r.context.GetStatus()
	var mutex sync.Mutex
	var leaderIDs []string
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

				if available, err := arangod.IsServerAvailable(ctxChild, c); err != nil {
					return err
				} else if !available {
					return errors.New("not available")
				}

				mutex.Lock()
				leaderIDs = append(leaderIDs, id)
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

	if len(leaderIDs) > 0 {
		return leaderIDs, nil
	}

	if anyError != nil {
		return nil, errors.WithMessagef(anyError, "unable to get a leader")
	}

	return nil, errors.New("unable to get a leader")
}

// setSingleServerLeadership adds or removes leadership label on a single server pod.
func (r *Resources) ensureSingleServerLeader(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	changed := false

	enabled := features.FailoverLeadership().Enabled()
	var leaderID string
	if enabled {
		leaderIDs, err := r.getSingleServerLeaderID(ctx)
		if err != nil {
			return err
		}

		if len(leaderIDs) == 1 {
			leaderID = leaderIDs[0]
		} else if len(leaderIDs) > 1 {
			r.log.Error("multiple leaders found: %s. Blocking traffic to the deployment services", strings.Join(leaderIDs, ", "))
		}
	}

	status := r.context.GetStatus()
	for _, m := range status.Members.Single {
		pod, exist := cachedStatus.Pod().V1().GetSimple(m.Pod.GetName())
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
			return errors.WithMessagef(err, "unable to change leader label for pod %s", m.Pod.GetName())
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
