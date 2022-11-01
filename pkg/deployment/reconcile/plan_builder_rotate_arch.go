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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createChangeMemberArchPlan goes over all pods to check if an Architecture type is set correctly
func (r *Reconciler) createChangeMemberArchPlan(ctx context.Context,
	apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	var p api.Plan

	for _, m := range status.Members.AsList() {
		member := m.Member
		cache, ok := context.ACS().ClusterCache(member.ClusterID)
		if !ok {
			return p
		}

		if pod, ok := cache.Pod().V1().GetSimple(member.Pod.GetName()); ok {
			if v, ok := pod.GetAnnotations()[deployment.ArangoDeploymentPodChangeArchAnnotation]; ok {
				arch := api.ArangoDeploymentArchitectureType(v)
				if arch.IsArchMismatch(spec.Architecture, member.Architecture) {
					if arch == api.ArangoDeploymentArchitectureARM64 && status.CurrentImage.ArangoDBVersion.CompareTo("3.10.0") < 0 {
						arch = api.ArangoDeploymentArchitectureAMD64
						r.log.Warn("Cannot apply ARM64 'arch' annotation. It's not supported in ArangoDB < 3.10.0")
						context.CreateEvent(k8sutil.NewCannotSetArchitectureARM64Event(apiObject, member.ID))
						context.CreateEvent(k8sutil.NewCannotSetArchitectureARM64Event(pod, member.ID))
						return p
					}

					r.log.
						Str("pod-name", member.Pod.GetName()).
						Str("server-group", m.Group.AsRole()).
						Warn("try changing an Architecture type, but %s", getRequiredRotateMessage(member.Pod.GetName()))
					p = append(p,
						actions.NewAction(api.ActionTypeSetCurrentMemberArch, m.Group, member, "Architecture Mismatch").SetArch(arch),
					)
				}
			}
		}
	}
	return p
}
