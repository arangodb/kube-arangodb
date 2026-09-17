//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	client "github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	sharedReconcile "github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (r *Reconciler) createMemberGatewayConfigConditionPlan(ctx context.Context, apiObject k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, planCtx PlanBuilderContext) api.Plan {
	var plan api.Plan

	// Check for members in failed state.
	for _, m := range status.Members.AsListInGroup(api.ServerGroupGateways) {
		hash, err := r.getGatewayMemberConfigHash(ctx, planCtx, apiObject, spec, m.Group, m.Member)
		if err != nil {
			r.log.Str("member", m.Member.ID).Err(err).Debug("Failed to get gateway config hash")
			if c, ok := m.Member.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || c.Status == core.ConditionTrue {
				plan = append(plan, sharedReconcile.UpdateMemberConditionActionV2("Config is not present", api.ConditionTypeGatewayConfig, m.Group, m.Member.ID, false, "Config is not present", "Config is not present", ""))
			}

			continue
		}

		r.log.Str("member", m.Member.ID).Str("hash", hash).Debug("Gateway config hash received")
		if c, ok := m.Member.Conditions.Get(api.ConditionTypeGatewayConfig); !ok || c.Status == core.ConditionFalse || c.Hash != hash {
			plan = append(plan, sharedReconcile.UpdateMemberConditionActionV2("Config Present", api.ConditionTypeGatewayConfig, m.Group, m.Member.ID, true, "Config Present", "Config Present", hash))
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

// createGatewayConfigPushPlan pushes the gateway dynamic config to gateway member sidecars over ADS when
// the deployment runs in push mode. The action itself is a no-op when a member already serves the desired
// revision, so it is safe to emit periodically (guarded by a back-off in the high plan).
func (r *Reconciler) createGatewayConfigPushPlan(_ context.Context, _ k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, _ PlanBuilderContext) api.Plan {
	if !spec.Gateway.IsDynamicModePush() {
		return nil
	}

	var plan api.Plan

	for _, m := range status.Members.AsListInGroup(api.ServerGroupGateways) {
		if m.Member.Phase != api.MemberPhaseCreated {
			continue
		}

		plan = append(plan, actions.NewAction(api.ActionTypeGatewayConfigPush, api.ServerGroupGateways, m.Member, "Push gateway config to member sidecar"))
	}

	return plan
}

// getGatewayMemberConfigHash returns the config revision the gateway member is currently serving. In push
// mode this is the version served by the sidecar ADS snapshot (queried over the EnvoyConfigV1 control
// channel); otherwise it is the hash reported by the gateway /_inventory endpoint. Keeping readiness in
// push mode tied to the ADS snapshot (rather than /_inventory, which reflects a different, mounted source)
// lets it converge to the pushed config.
func (r *Reconciler) getGatewayMemberConfigHash(ctx context.Context, planCtx PlanBuilderContext, apiObject k8sutil.APIObject, spec api.DeploymentSpec, group api.ServerGroup, member api.MemberStatus) (string, error) {
	if spec.Gateway.IsDynamicModePush() {
		return r.getGatewayPushedVersion(ctx, planCtx, apiObject, spec, member)
	}

	inv, err := r.getGatewayInventoryConfig(ctx, planCtx, group, member)
	if err != nil {
		return "", err
	}

	return inv.Configuration.Hash, nil
}

// getGatewayPushedVersion queries the member sidecar's EnvoyConfigV1 control channel for the version of the
// config it is currently serving over ADS.
func (r *Reconciler) getGatewayPushedVersion(ctx context.Context, planCtx PlanBuilderContext, apiObject k8sutil.APIObject, spec api.DeploymentSpec, member api.MemberStatus) (string, error) {
	lCtx, c := context.WithTimeout(ctx, 5*time.Second)
	defer c()

	token, err := planCtx.GetMembersToken(lCtx)
	if err != nil {
		return "", err
	}

	cl, closer, err := newGatewayConfigClient(lCtx, apiObject, spec, member.ID, token)
	if err != nil {
		return "", err
	}
	defer closer()

	st, err := cl.Status(lCtx, &pbSharedV1.Empty{})
	if err != nil {
		return "", err
	}

	return st.GetVersion(), nil
}

func (r *Reconciler) getGatewayInventoryConfig(ctx context.Context, planCtx PlanBuilderContext, group api.ServerGroup, member api.MemberStatus) (client.Inventory, error) {
	// Use GetServerClient to get a validated connection. The timeout must
	// be large enough to cover both the internal Version() health check
	// in GetServerClient and the Inventory() call itself, which goes
	// through Envoy's ext_authz filter.
	lCtx, c := context.WithTimeout(ctx, 5*time.Second)
	defer c()

	serverClient, err := planCtx.GetServerClient(lCtx, group, member.ID)
	if err != nil {
		r.log.Str("member", member.ID).Err(err).Debug("Failed to get server client for gateway inventory")
		return client.Inventory{}, err
	}

	internalClient := client.NewClient(serverClient.Connection())

	inv, err := internalClient.Inventory(lCtx)
	if err != nil {
		r.log.Str("member", member.ID).Err(err).Debug("Failed to fetch gateway inventory")
		return client.Inventory{}, err
	}

	return inv, nil
}

func (r *Reconciler) createGatewaySidecarEnablementPlan(ctx context.Context, _ k8sutil.APIObject, spec api.DeploymentSpec,
	status api.DeploymentStatus, planCtx PlanBuilderContext) api.Plan {
	expected := features.GatewayIntegration().ImageSupported(status.CurrentImage) && spec.Sidecar.IsEnabled(spec.IsGatewayEnabled())

	if expected {
		if !status.Conditions.IsTrue(api.ConditionTypeGatewaySidecarEnabled) {
			return api.Plan{
				sharedReconcile.UpdateConditionActionV2("Gateways Sidecar Enabled", api.ConditionTypeGatewaySidecarEnabled, true, "Gateway Enabled", "Gateway Enabled", ""),
				sharedReconcile.UpdateConditionActionV2("Gateways Sidecar Enabled", api.ConditionTypeProfilesReady, false, "Gateway Enabled", "Gateway Enabled", ""),
				sharedReconcile.UpdateConditionActionV2("Gateways Sidecar Enabled", api.ConditionTypeStorageReady, false, "Gateway Enabled", "Gateway Enabled", ""),
			}
		}
	} else {
		if status.Conditions.IsTrue(api.ConditionTypeGatewaySidecarEnabled) {
			return api.Plan{
				sharedReconcile.RemoveConditionActionV2("Gateways Sidecar Disabled", api.ConditionTypeGatewaySidecarEnabled),
				sharedReconcile.UpdateConditionActionV2("Gateways Sidecar Disabled", api.ConditionTypeProfilesReady, false, "Gateway Disabled", "Gateway Disabled", ""),
				sharedReconcile.UpdateConditionActionV2("Gateways Sidecar Disabled", api.ConditionTypeStorageReady, false, "Gateway Disabled", "Gateway Disabled", ""),
			}
		}
	}

	return nil
}
