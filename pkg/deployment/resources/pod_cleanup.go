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
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/acs/sutil"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/info"
	podv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

const (
	statelessTerminationPeriod         = time.Minute                    // We wait this long for a stateless server to terminate on it's own. Afterwards we kill it.
	recheckStatefullPodCleanupInterval = util.Interval(time.Second * 2) // Interval used when Pod finalizers need to be rechecked soon
)

// CleanupTerminatedPods removes all pods in Terminated state that belong to a member in Created state.
// Returns: Interval_till_next_inspection, error
func (r *Resources) CleanupTerminatedPods(ctx context.Context) (util.Interval, error) {
	log := r.log.Str("section", "pod")
	nextInterval := maxPodInspectorInterval // Large by default, will be made smaller if needed in the rest of the function

	// Update member status from all pods found
	status := r.context.GetStatus()
	if err := r.context.ACS().ForEachHealthyCluster(func(item sutil.ACSItem) error {
		return item.Cache().Pod().V1().Iterate(func(pod *core.Pod) error {
			if info.GetPodServerGroup(pod) == api.ServerGroupImageDiscovery {
				// Image ID pods are not relevant to inspect here
				return nil
			}

			// Find member status
			memberStatus, group, found := status.Members.MemberStatusByPodName(pod.GetName())
			if !found {
				log.Str("pod", pod.GetName()).Debug("no memberstatus found for pod. Performing cleanup")
			} else {
				spec := r.context.GetSpec()
				coreContainers := spec.GetCoreContainers(group)
				if !(k8sutil.IsPodSucceeded(pod, coreContainers) || k8sutil.IsPodFailed(pod, coreContainers) ||
					k8sutil.IsPodTerminating(pod)) {
					// The pod is not being terminated or failed or succeeded.
					return nil
				}

				// Check member termination condition
				if !memberStatus.Conditions.IsTrue(api.ConditionTypeTerminated) {
					if !group.IsStateless() {
						// For statefull members, we have to wait for confirmed termination
						log.Str("pod", pod.GetName()).Debug("Cannot cleanup pod yet, waiting for it to reach terminated state")
						nextInterval = nextInterval.ReduceTo(recheckStatefullPodCleanupInterval)
						return nil
					} else {
						// If a stateless server does not terminate within a reasonable amount or time, we kill it.
						t := pod.GetDeletionTimestamp()
						if t == nil || t.Add(statelessTerminationPeriod).After(time.Now()) {
							// Either delete timestamp is not set, or not yet waiting long enough
							nextInterval = nextInterval.ReduceTo(util.Interval(statelessTerminationPeriod))
							return nil
						}
					}
				}
			}

			// Ok, we can delete the pod
			log.Str("pod-name", pod.GetName()).Debug("Cleanup terminated pod")

			options := meta.NewDeleteOptions(0)
			options.Preconditions = meta.NewUIDPreconditions(string(pod.GetUID()))
			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				return item.Cache().Client().Kubernetes().CoreV1().Pods(item.Cache().Namespace()).Delete(ctxChild, pod.GetName(), *options)
			})
			if err != nil && !kerrors.IsNotFound(err) {
				log.Err(err).Str("pod", pod.GetName()).Debug("Failed to cleanup pod")
				return errors.WithStack(err)
			}

			return nil
		}, podv1.FilterPodsByLabels(k8sutil.LabelsForDeployment(r.context.GetAPIObject().GetName(), "")))
	}); err != nil {
		return 0, err
	}

	return nextInterval, nil
}
