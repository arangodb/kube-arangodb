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

package deployment

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/names"
)

func (d *Deployment) createInitialTopology(ctx context.Context) error {
	spec := d.GetSpec()
	return d.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		if spec.Topology.IsEnabled() {
			if s.Topology == nil {
				s.Topology = api.NewTopologyStatus(spec.Topology)
				return true
			}
		} else {
			if s.Topology != nil {
				s.Topology = nil
				return true
			}
		}
		return false
	})
}

func (d *Deployment) renderMemberID(spec api.DeploymentSpec, status *api.DeploymentStatus, groupStatus *api.ServerGroupStatus, group api.ServerGroup) string {
	switch spec.GetServerGroupSpec(group).IndexMethod.Get() {
	case api.ServerGroupIndexMethodOrdered:
		if v := groupStatus.Index; v == nil {
			groupStatus.Index = util.NewType[int](0)
			return names.GetArangodIDInt(group, 0)
		} else {
			z := *v
			z++
			groupStatus.Index = util.NewType[int](z)
			return names.GetArangodIDInt(group, z)
		}
	default:
		for {
			if id := names.GetArangodID(group); !status.Members.ContainsID(id) {
				return id
			}
		}
	}
}
