//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
//

package resources

import (
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

// CleanupTerminatedPods removes all pods in Terminated state that belong to a member in Created state.
func (r *Resources) CleanupTerminatedPods() error {
	log := r.log

	pods, err := r.context.GetOwnedPods()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get owned pods")
		return maskAny(err)
	}

	// Update member status from all pods found
	status := r.context.GetStatus()
	for _, p := range pods {
		if k8sutil.IsArangoDBImageIDAndVersionPod(p) {
			// Image ID pods are not relevant to inspect here
			continue
		}

		// Check pod state
		if !(k8sutil.IsPodSucceeded(&p) || k8sutil.IsPodFailed(&p)) {
			continue
		}

		// Find member status
		memberStatus, _, found := status.Members.MemberStatusByPodName(p.GetName())
		if !found {
			log.Debug().Str("pod", p.GetName()).Msg("no memberstatus found for pod")
			continue
		}

		// Check member termination condition
		if !memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
			continue
		}

		// Ok, we can delete the pod
		log.Debug().Str("pod-name", p.GetName()).Msg("Cleanup terminated pod")
		if err := r.context.CleanupPod(p); err != nil {
			log.Warn().Err(err).Str("pod-name", p.GetName()).Msg("Failed to cleanup pod")
		}
	}
	return nil
}
