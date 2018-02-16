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

package deployment

import (
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
)

// inspectPods lists all pods that belong to the given deployment and updates
// the member status of the deployment accordingly.
func (d *Deployment) inspectPods() error {
	log := d.deps.Log

	log.Debug().Msg("inspecting pods")
	pods, err := d.deps.KubeCli.CoreV1().Pods(d.apiObject.GetNamespace()).List(k8sutil.DeploymentListOpt(d.apiObject.GetName()))
	if err != nil {
		log.Debug().Err(err).Msg("Failed to list pods")
		return maskAny(err)
	}

	// Update member status from all pods found
	for _, p := range pods.Items {
		// Check ownership
		if !d.isOwnerOf(&p) {
			continue
		}

		// Find member status
		memberStatus, group, found := d.status.Members.MemberStatusByPodName(p.GetName())
		if !found {
			continue
		}

		// Update state
		log.Debug().Str("pod-name", p.GetName()).Msg("found member status for pod")
		if memberStatus.State == api.MemberStateCreating {
			if k8sutil.IsPodReady(&p) {
				memberStatus.State = api.MemberStateReady
				if err := d.status.Members.UpdateMemberStatus(memberStatus, group); err != nil {
					return maskAny(err)
				}
				log.Debug().Str("pod-name", p.GetName()).Msg("updated member status for pod to ready")
			}
		}
	}

	// Save status
	if err := d.updateCRStatus(); err != nil {
		return maskAny(err)
	}
	return nil
}
