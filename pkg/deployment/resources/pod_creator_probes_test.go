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

package resources

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

// setGatewayTLSFeature toggles the gateway-tls operator feature for the duration of a test and
// restores the previous value.
func setGatewayTLSFeature(t *testing.T, enabled bool) {
	p := features.GatewayTLS().EnabledPointer()
	restore := *p
	t.Cleanup(func() { *p = restore })
	*p = enabled
}

func specCentralTLS(secure bool) api.DeploymentSpec {
	var s api.DeploymentSpec
	if !secure {
		// CASecretNameDisabled ("None") turns the deployment (arangod) TLS off.
		s.TLS.CASecretName = util.NewType(api.CASecretNameDisabled)
	}
	return s
}

func withGatewayTLS(s api.DeploymentSpec, enabled bool) api.DeploymentSpec {
	if s.Gateway == nil {
		s.Gateway = &api.DeploymentSpecGateway{}
	}
	s.Gateway.TLS = &api.DeploymentSpecGatewayTLS{Enabled: util.NewType(enabled)}
	return s
}

// Test_probeIsSecure verifies that the probe scheme follows the deployment (arangod) TLS for every
// group, and additionally follows the gateway's client-facing TLS for the gateway group when the
// gateway-tls feature decouples it (spec.gateway.tls) while arangod TLS is disabled.
func Test_probeIsSecure(t *testing.T) {
	// Sanity: default (empty) TLS spec is secure, CASecretNameDisabled is not.
	require.True(t, specCentralTLS(true).TLS.IsSecure())
	require.False(t, specCentralTLS(false).TLS.IsSecure())

	t.Run("Central arangod TLS drives all groups", func(t *testing.T) {
		setGatewayTLSFeature(t, false)

		for _, g := range []api.ServerGroup{api.ServerGroupAgents, api.ServerGroupDBServers,
			api.ServerGroupCoordinators, api.ServerGroupSingle, api.ServerGroupGateways} {
			require.True(t, probeIsSecure(specCentralTLS(true), g), "group %s secure when arangod TLS on", g.AsRole())
			require.False(t, probeIsSecure(specCentralTLS(false), g), "group %s plaintext when arangod TLS off", g.AsRole())
		}
	})

	t.Run("Decoupled: gateway TLS on, arangod TLS off", func(t *testing.T) {
		setGatewayTLSFeature(t, true)
		spec := withGatewayTLS(specCentralTLS(false), true)

		// The gateway serves client-facing HTTPS, so its probe must be secure even though arangod
		// TLS is disabled - this is the case the CI e2e caught (probe stayed HTTP, pod never Ready).
		require.True(t, probeIsSecure(spec, api.ServerGroupGateways),
			"gateway probe must be secure when gateway serves client-facing TLS")

		// Non-gateway groups still follow arangod TLS (disabled here).
		require.False(t, probeIsSecure(spec, api.ServerGroupCoordinators),
			"arangod groups stay plaintext when only the gateway serves TLS")
		require.False(t, probeIsSecure(spec, api.ServerGroupDBServers))
	})

	t.Run("Gateway TLS requested but feature disabled", func(t *testing.T) {
		setGatewayTLSFeature(t, false)
		spec := withGatewayTLS(specCentralTLS(false), true)

		require.False(t, probeIsSecure(spec, api.ServerGroupGateways),
			"spec.gateway.tls has no effect while the gateway-tls feature is disabled")
	})

	t.Run("Feature enabled but spec.gateway.tls disabled", func(t *testing.T) {
		setGatewayTLSFeature(t, true)
		spec := withGatewayTLS(specCentralTLS(false), false)

		require.False(t, probeIsSecure(spec, api.ServerGroupGateways),
			"gateway probe stays plaintext when spec.gateway.tls is not enabled")
	})

	t.Run("Central TLS on keeps gateway secure regardless of feature", func(t *testing.T) {
		setGatewayTLSFeature(t, false)
		require.True(t, probeIsSecure(specCentralTLS(true), api.ServerGroupGateways))
	})
}
