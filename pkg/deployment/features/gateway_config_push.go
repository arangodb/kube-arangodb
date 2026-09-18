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

package features

import api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

func init() {
	registerFeature(gatewayConfigPush)
}

var gatewayConfigPush = &feature{
	name:               "gateway-config-push",
	description:        "Defines if the operator delivers the dynamic gateway config by pushing it to the gateway sidecar over ADS (default) instead of relying on the mounted ConfigMap",
	enterpriseRequired: false,
	enabledByDefault:   true,
	hidden:             true,
}

// GatewayConfigPush returns the feature gating operator push (ADS) delivery of the dynamic gateway config.
func GatewayConfigPush() Feature {
	return gatewayConfigPush
}

// GatewayDynamicModePush resolves whether the dynamic gateway config should be delivered in push (ADS)
// mode for the given gateway spec. Push is the default while the feature is enabled; an explicit
// spec.gateway.dynamicMode wins (push or configmap), and the feature being disabled forces ConfigMap
// delivery regardless of the spec.
func GatewayDynamicModePush(g *api.DeploymentSpecGateway) bool {
	if !g.IsDynamic() || !gatewayConfigPush.Enabled() {
		return false
	}

	if g.DynamicMode != nil {
		return *g.DynamicMode == api.GatewayDynamicModePush
	}

	return true
}
