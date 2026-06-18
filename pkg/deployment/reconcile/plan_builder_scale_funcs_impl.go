//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func planBuilderScaleDownTopologyAwarenessMember(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatus, bool, error) {
	if !status.Topology.Enabled() {
		return api.MemberStatus{}, false, nil
	}

	r, max := -1, -1

	for id, z := range status.Topology.Zones {
		if v := len(z.Members); v > max {
			r = id
			max = v
		}
	}

	if r == -1 {
		return api.MemberStatus{}, false, nil
	}

	z := status.Topology.Zones[r]

	members := z.Get(group)

	perm := util.Rand().Perm(len(members))
	for _, idx := range perm {
		mid := members[idx]

		member, ok := in.ElementByID(mid)
		if !ok {
			continue
		}

		if member.Phase == api.MemberPhaseCreated {
			return member, true, nil
		}
	}

	return api.MemberStatus{}, false, nil
}

func planBuilderScaleDownTopologyMissing(context PlanBuilderContext, status api.DeploymentStatus, group api.ServerGroup, in api.MemberStatusList) (api.MemberStatus, bool, error) {
	if !status.Topology.Enabled() {
		return api.MemberStatus{}, false, nil
	}

	for _, el := range in {
		if el.Topology == nil || el.Topology.ID != status.Topology.GetID() {
			return el, true, nil
		}
	}

	return api.MemberStatus{}, false, nil
}
