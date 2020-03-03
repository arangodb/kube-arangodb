//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

const (
	statelessTerminationPeriod         = time.Minute                    // We wait this long for a stateless server to terminate on it's own. Afterwards we kill it.
	recheckStatefullPodCleanupInterval = util.Interval(time.Second * 2) // Interval used when Pod finalizers need to be rechecked soon
)

// CleanupTerminatedPods removes all pods in Terminated state that belong to a member in Created state.
// Returns: Interval_till_next_inspection, error
func (r *Resources) CleanupTerminatedPods() (util.Interval, error) {
	log := r.log
	nextInterval := maxPodInspectorInterval // Large by default, will be made smaller if needed in the rest of the function

	pods, err := r.context.GetOwnedPods()
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get owned pods")
		return 0, maskAny(err)
	}

	// Update member status from all pods found
	status, _ := r.context.GetStatus()
	for _, p := range pods {
		if k8sutil.IsArangoDBImageIDAndVersionPod(p) {
			// Image ID pods are not relevant to inspect here
			continue
		}

		// Check pod state
		if !(k8sutil.IsPodSucceeded(&p) || k8sutil.IsPodFailed(&p) || k8sutil.IsPodTerminating(&p)) {
			continue
		}

		// Find member status
		memberStatus, group, found := status.Members.MemberStatusByPodName(p.GetName())
		if !found {
			log.Debug().Str("pod", p.GetName()).Msg("no memberstatus found for pod. Performing cleanup")
		} else {
			// Check member termination condition
			if !memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
				if !group.IsStateless() {
					// For statefull members, we have to wait for confirmed termination
					log.Debug().Str("pod", p.GetName()).Msg("Cannot cleanup pod yet, waiting for it to reach terminated state")
					nextInterval = nextInterval.ReduceTo(recheckStatefullPodCleanupInterval)
					continue
				} else {
					// If a stateless server does not terminate within a reasonable amount or time, we kill it.
					t := p.GetDeletionTimestamp()
					if t == nil || t.Add(statelessTerminationPeriod).After(time.Now()) {
						// Either delete timestamp is not set, or not yet waiting long enough
						nextInterval = nextInterval.ReduceTo(util.Interval(statelessTerminationPeriod))
						continue
					}
				}
			}
		}

		// Ok, we can delete the pod
		log.Debug().Str("pod-name", p.GetName()).Msg("Cleanup terminated pod")
		if err := r.context.CleanupPod(p); err != nil {
			log.Warn().Err(err).Str("pod-name", p.GetName()).Msg("Failed to cleanup pod")
		}
	}
	return nextInterval, nil
}
