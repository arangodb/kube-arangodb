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

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	client "github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	sharedReconcile "github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
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

		if c, ok := m.Member.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || c.Status == core.ConditionFalse || c.Hash != inv.Configuration.Hash {
			plan = append(plan, sharedReconcile.UpdateMemberConditionActionV2("Config Present", api.ConditionTypeGatewayConfig, m.Group, m.Member.ID, true, "Config Present", "Config Present", inv.Configuration.Hash))
		}
	}

	return plan
}

func (r *Reconciler) createGatewayConfigConditionPlan(ctx context.Context, _ k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, planCtx PlanBuilderContext) api.Plan {
	var plan api.Plan

	if spec.Gateway.IsEnabled() {
		cm, exists := planCtx.ACS().CurrentClusterCache().ConfigMap().V1().GetSimple(resources.GetGatewayConfigMapName(r.context.GetAPIObject().GetName()))
		if !exists {
			if c, ok := status.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || c.Status == core.ConditionTrue || c.Hash != "" {
				plan = append(plan, sharedReconcile.UpdateConditionActionV2("Gateway CM Missing", api.ConditionTypeGatewayConfig, false, "Gateway CM Missing", "Gateway CM Missing", ""))
			}
			return plan
		}

		if cm == nil || cm.Data[utilConstants.GatewayConfigChecksum] == "" {
			if c, ok := status.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || c.Status == core.ConditionTrue || c.Hash != "" {
				plan = append(plan, sharedReconcile.UpdateConditionActionV2("Gateway CM Missing", api.ConditionTypeGatewayConfig, false, "Gateway CM Missing", "Gateway CM Missing", ""))
			}
			return plan
		}

		checksum := cm.Data[utilConstants.GatewayConfigChecksum]

		cok := true
		for _, m := range status.Members.AsListInGroup(api.ServerGroupGateways) {
			if v, ok := m.Member.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || v.Status != core.ConditionTrue || v.Hash != checksum {
				cok = false
			}
			if !cok {
				break
			}
		}

		if cok {
			if c, ok := status.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || c.Status == core.ConditionFalse || c.Hash != checksum {
				plan = append(plan, sharedReconcile.UpdateConditionActionV2("Gateway Config UpToDate", api.ConditionTypeGatewayConfig, true, "Gateway Config Propagated", "Gateway Config Propagated", checksum))
				return plan
			}
		} else {
			if c, ok := status.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || c.Status == core.ConditionTrue || c.Hash != checksum {
				plan = append(plan, sharedReconcile.UpdateConditionActionV2("Gateway Config Not UpToDate", api.ConditionTypeGatewayConfig, false, "Gateway Config Not Propagated", "Gateway Config Not Propagated", checksum))
				return plan
			}
		}

	} else {
		if _, ok := status.Conditions.Get(api.ConditionTypeGatewayConfig); ok {
			plan = append(plan, sharedReconcile.RemoveConditionActionV2("Gateways Disabled", api.ConditionTypeGatewayConfig))
			return plan
		}
	}

	return plan
}

func (r *Reconciler) getGatewayInventoryConfig(ctx context.Context, planCtx PlanBuilderContext, group api.ServerGroup, member api.MemberStatus) (*client.Inventory, error) {
	serverClient, err := planCtx.GetServerClient(ctx, group, member.ID)
	if err != nil {
		return nil, err
	}

	internalClient := client.NewClient(serverClient.Connection(), logger)

	lCtx, c := context.WithTimeout(ctx, 500*time.Millisecond)
	defer c()

	return internalClient.Inventory(lCtx)
}

func (r *Reconciler) createGatewaySidecarEnablementPlan(ctx context.Context, _ k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, planCtx PlanBuilderContext) api.Plan {
	expected := features.GatewayIntegration().ImageSupported(status.CurrentImage) && spec.Sidecar.IsEnabled(spec.IsGatewayEnabled())

	if expected {
		if !status.Conditions.IsTrue(api.ConditionTypeGatewaySidecarEnabled) {
			return api.Plan{sharedReconcile.UpdateConditionActionV2("Gateways Sidecar Enabled", api.ConditionTypeGatewaySidecarEnabled, true, "Gateway Enabled", "Gateway Enabled", "")}
		}
	} else {
		if status.Conditions.IsTrue(api.ConditionTypeGatewaySidecarEnabled) {
			return api.Plan{sharedReconcile.RemoveConditionActionV2("Gateways Sidecar Disabled", api.ConditionTypeGatewaySidecarEnabled)}
		}
	}

	return nil
}
