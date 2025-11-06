//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
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
	if !r.context.GetSpec().GetMode().HasAgents() {
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
		if _, c, err := patcher.Patcher[*core.Service](ctx, cachedStatus.ServicesModInterface().V1(), s, meta.PatchOptions{}, patcher.PatchServiceSelector(selector), patcher.PatchServicePorts(ports)); err != nil {
			return err
		} else {
			if !c {
				return nil
			}

			return errors.Reconcile()
		}
	}

	s := r.createService(leaderAgentSvcName, r.context.GetNamespace(), "", core.ServiceTypeClusterIP, true, r.context.GetAPIObject().AsOwner(), ports, selector)
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
