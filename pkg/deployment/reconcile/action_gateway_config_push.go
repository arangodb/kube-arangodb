//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

	discoveryApi "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	pbEnvoyConfigV1 "github.com/arangodb/kube-arangodb/integrations/envoy/config/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

func newGatewayConfigPushAction(action api.Action, actionCtx ActionContext) Action {
	a := &actionGatewayConfigPush{}

	a.actionImpl = newActionImplDefRef(action, actionCtx)

	return a
}

type actionGatewayConfigPush struct {
	actionImpl

	actionEmptyCheckProgress
}

// Start pushes the current gateway dynamic config (CDS/LDS) to the member sidecar over the EnvoyConfigV1
// control channel, but only if the sidecar is serving a different revision. It only acts in push mode.
func (a *actionGatewayConfigPush) Start(ctx context.Context) (bool, error) {
	spec := a.actionCtx.GetSpec()

	if !features.GatewayDynamicModePush(spec.Gateway) {
		return true, nil
	}

	m, _, ok := a.actionCtx.GetStatus().Members.ElementByID(a.MemberID())
	if !ok {
		a.log.Warn("Member not found")
		return true, nil
	}

	cache := a.actionCtx.ACS().CurrentClusterCache()
	name := a.actionCtx.GetName()

	// The desired CDS/LDS and revision (checksum) are rendered into the gateway ConfigMaps already.
	cds, cok := cache.ConfigMap().V1().GetSimple(resources.GetGatewayConfigMapName(name, "cds"))
	lds, lok := cache.ConfigMap().V1().GetSimple(resources.GetGatewayConfigMapName(name, "lds"))
	if !cok || !lok || cds == nil || lds == nil {
		a.log.Debug("Gateway CDS/LDS config maps are not present yet")
		return true, nil
	}

	version := cds.Data[utilConstants.GatewayConfigChecksum]
	if version == "" {
		a.log.Debug("Gateway config checksum is not present yet")
		return true, nil
	}

	cdsResp, err := ugrpc.UnmarshalYAML[*discoveryApi.DiscoveryResponse]([]byte(cds.Data[utilConstants.GatewayConfigFileName]))
	if err != nil {
		return false, err
	}

	ldsResp, err := ugrpc.UnmarshalYAML[*discoveryApi.DiscoveryResponse]([]byte(lds.Data[utilConstants.GatewayConfigFileName]))
	if err != nil {
		return false, err
	}

	token, err := a.actionCtx.GetMembersToken(ctx)
	if err != nil {
		return false, err
	}

	client, closer, err := newGatewayConfigClient(ctx, a.actionCtx.GetAPIObject(), spec, m.ID, token)
	if err != nil {
		return false, err
	}
	defer closer()

	// Skip the push when the sidecar already serves the desired revision.
	if st, err := client.Status(ctx, &pbSharedV1.Empty{}); err == nil && st.GetVersion() == version {
		return true, nil
	}

	if _, err := client.Push(ctx, &pbEnvoyConfigV1.EnvoyConfigV1PushRequest{
		Version:   version,
		Clusters:  cdsResp.GetResources(),
		Listeners: ldsResp.GetResources(),
	}); err != nil {
		return false, err
	}

	a.log.Str("member", m.ID).Str("version", version).Info("Pushed gateway config to member sidecar")

	return true, nil
}
