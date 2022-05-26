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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
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

	leaderID := r.context.GetAgencyLeaderID()
	status, _ := r.context.GetStatus()
	noLeader := len(leaderID) == 0
	changed := false
	group := api.ServerGroupAgents
	agencyServers := func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			pod, exist := cachedStatus.Pod().V1().GetSimple(m.PodName)
			if !exist {
				continue
			}

			labels := pod.GetLabels()
			if noLeader || m.ID != leaderID {
				// Unset a leader when:
				// - leader is unknown.
				// - leader does not belong to the current pod.

				if _, ok := labels[k8sutil.LabelKeyArangoLeader]; ok {
					delete(labels, k8sutil.LabelKeyArangoLeader)

					err := r.context.ApplyPatchOnPod(ctx, pod, patch.ItemReplace(patch.NewPath("metadata", "labels"), labels))
					if err != nil {
						r.log.Error().Err(err).Msgf("Unable to remove leader label")
						return err
					}

					r.log.Info().Msgf("leader label is removed from \"%s\" member", m.ID)
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
				r.log.Error().Err(err).Msgf("Unable to update leader label")
				return err
			}
			r.log.Info().Msgf("leader label is set on \"%s\" member", m.ID)
			changed = true
		}

		return nil
	}

	if err := status.Members.ForeachServerInGroups(agencyServers, group); err != nil {
		return err
	}

	if changed {
		return errors.Reconcile()
	}
	changed = false

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
				// The service is not changed.
				return nil
			}

			return errors.Reconcile()
		} else {
			return err
		}
	}

	s := r.createService(leaderAgentSvcName, r.context.GetNamespace(), r.context.GetAPIObject().AsOwner(), shared.ArangoPort, selector)
	err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		_, err := r.context.ServicesModInterface().Create(ctxChild, s, meta.CreateOptions{})
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
