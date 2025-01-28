//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package route

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func (h *handler) HandleArangoDestination(ctx context.Context, item operation.Item, extension *networkingApi.ArangoRoute, status *networkingApi.ArangoRouteStatus, deployment *api.ArangoDeployment) (*operator.Condition, bool, error) {
	if dest := extension.Spec.GetDestination(); dest != nil {
		if svc := dest.GetService(); svc != nil {
			return h.HandleArangoDestinationService(ctx, item, extension, status, deployment, dest, svc)
		}
		if endpoints := dest.GetEndpoints(); endpoints != nil {
			return h.HandleArangoDestinationEndpoints(ctx, item, extension, status, deployment, dest, endpoints)
		}
	}

	return &operator.Condition{
		Status:  false,
		Reason:  "Destination Not Found",
		Message: "Destination Not Found",
	}, false, nil
}

func (h *handler) HandleArangoDestinationWithTargets(ctx context.Context, item operation.Item, extension *networkingApi.ArangoRoute, status *networkingApi.ArangoRouteStatus, depl *api.ArangoDeployment) (*operator.Condition, bool, error) {
	c, changed, err := h.HandleArangoDestination(ctx, item, extension, status, depl)

	if operator.IsTemporary(err) {
		return nil, false, err
	}

	if c == nil && !c.Status && status.Target != nil {
		status.Target = nil
		changed = true
	}

	return c, changed, err
}

func (h *handler) HandleDestinationRequired(ctx context.Context, item operation.Item, extension *networkingApi.ArangoRoute, status *networkingApi.ArangoRouteStatus, _ *api.ArangoDeployment) (bool, error) {
	if !status.Conditions.IsTrue(networkingApi.DestinationValidCondition) {
		return false, operator.Stop("Destination is not ready")
	}

	return false, nil
}
