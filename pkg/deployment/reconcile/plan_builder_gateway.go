//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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
	"time"

	core "k8s.io/api/core/v1"

	pbInventoryV1 "github.com/arangodb/kube-arangodb/integrations/inventory/v1/definition"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	client "github.com/arangodb/kube-arangodb/pkg/deployment/client"
	sharedReconcile "github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (r *Reconciler) createMemberGatewayConfigConditionPlan(ctx context.Context, _ k8sutil.APIObject, _ api.DeploymentSpec,
	status api.DeploymentStatus, planCtx PlanBuilderContext) api.Plan {
	var plan api.Plan

	// Check for members in failed state.
	for _, m := range status.Members.AsListInGroup(api.ServerGroupGateways) {
		inv, err := r.getGatewayInventoryConfig(ctx, planCtx, m.Group, m.Member)
		if err != nil {
			if c, ok := m.Member.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || c.Status == core.ConditionTrue {
				plan = append(plan, sharedReconcile.UpdateMemberConditionActionV2("Config is not present", api.ConditionTypeGatewayConfig, m.Group, m.Member.ID, false, "Config is not present", "Config is not present", ""))
			}

			continue
		}

		if c, ok := m.Member.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || c.Status == core.ConditionFalse || c.Hash != inv.GetConfiguration().GetHash() {
			plan = append(plan, sharedReconcile.UpdateMemberConditionActionV2("Config Present", api.ConditionTypeGatewayConfig, m.Group, m.Member.ID, false, "Config Present", "Config Present", inv.GetConfiguration().GetHash()))
		}
	}

	return plan
}

func (r *Reconciler) getGatewayInventoryConfig(ctx context.Context, planCtx PlanBuilderContext, group api.ServerGroup, member api.MemberStatus) (*pbInventoryV1.Inventory, error) {
	serverClient, err := planCtx.GetServerClient(ctx, group, member.ID)
	if err != nil {
		return nil, err
	}

	internalClient := client.NewClient(serverClient.Connection(), logger)

	lCtx, c := context.WithTimeout(ctx, 500*time.Millisecond)
	defer c()

	return internalClient.Inventory(lCtx)
}
