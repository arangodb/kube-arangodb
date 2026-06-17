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
	"context"

	"k8s.io/apimachinery/pkg/api/equality"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

type actionTopologyZonesUpdate struct {
	actionImpl

	actionEmptyCheckProgress
}

func (t actionTopologyZonesUpdate) Start(ctx context.Context) (bool, error) {
	status := t.actionCtx.GetStatus()
	spec := t.actionCtx.GetSpec()

	if !status.Topology.Enabled() {
		return true, nil
	}

	if spec.GetMode() == api.DeploymentModeSingle {
		// Topology cannot be changed in single server deployments
		return true, nil
	}

	mapping := getTopologyMappingObject(status)

	if err := t.actionCtx.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		if !equality.Semantic.DeepEqual(s.Topology.Zones, mapping) {
			s.Topology.Zones = mapping
			return true
		}
		return false
	}); err != nil {
		t.log.Err(err).Error("Unable to propagate state of Status")
		return true, nil
	}

	return true, nil
}
