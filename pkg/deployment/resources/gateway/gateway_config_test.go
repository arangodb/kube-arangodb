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

package gateway

import (
	"fmt"
	"testing"

	bootstrapAPI "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	httpConnectionManagerAPI "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func renderAndPrintGatewayConfig(t *testing.T, cfg Config, validates ...func(t *testing.T, b *bootstrapAPI.Bootstrap)) {
	require.NoError(t, cfg.Validate())

	data, checksum, obj, err := cfg.RenderYAML()
	require.NoError(t, err)

	t.Logf("Checksum: %s", checksum)
	t.Log(string(data))

	for id := range validates {
		t.Run(fmt.Sprintf("Validation%d", id), func(t *testing.T) {
			validates[id](t, obj)
		})
	}
}

func Test_GatewayConfig(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
			},
		}, func(t *testing.T, b *bootstrapAPI.Bootstrap) {
			require.NotNil(t, b)
			require.NotNil(t, b.StaticResources)
			require.NotNil(t, b.StaticResources.Clusters)
			require.Len(t, b.StaticResources.Clusters, 1)
			require.NotNil(t, b.StaticResources.Clusters[0])
			require.NotNil(t, b.StaticResources.Clusters[0].LoadAssignment)
			require.NotNil(t, b.StaticResources.Clusters[0].LoadAssignment.Endpoints)
			require.Len(t, b.StaticResources.Clusters[0].LoadAssignment.Endpoints, 1)
			require.NotNil(t, b.StaticResources.Clusters[0].LoadAssignment.Endpoints[0])
			require.NotNil(t, b.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints)
			require.Len(t, b.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints, 1)
			require.NotNil(t, b.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0])
			require.NotNil(t, b.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0].GetEndpoint())
			require.NotNil(t, b.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0].GetEndpoint().Address)
			require.NotNil(t, b.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetSocketAddress())
			require.EqualValues(t, "127.0.0.1", b.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetSocketAddress().Address)
			require.EqualValues(t, 12345, b.StaticResources.Clusters[0].LoadAssignment.Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetSocketAddress().GetPortValue())
		})
	})
	t.Run("Without WebSocket", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
			},
		}, func(t *testing.T, b *bootstrapAPI.Bootstrap) {
			require.NotNil(t, b)
			require.NotNil(t, b.StaticResources)
			require.NotNil(t, b.StaticResources.Listeners)
			require.Len(t, b.StaticResources.Listeners, 1)
			require.NotNil(t, b.StaticResources.Listeners[0])
			require.NotNil(t, b.StaticResources.Listeners[0].DefaultFilterChain)
			require.NotNil(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters)
			require.Len(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters, 1)
			require.NotNil(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters[0])
			var o httpConnectionManagerAPI.HttpConnectionManager
			tgrpc.GRPCAnyCastAs(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters[0].GetTypedConfig(), &o)
			rc := o.GetRouteConfig()
			require.NotNil(t, rc)
			require.NotNil(t, rc.VirtualHosts)
			require.Len(t, rc.VirtualHosts, 1)
			require.NotNil(t, rc.VirtualHosts[0])
			require.Len(t, rc.VirtualHosts[0].Routes, 1)
			require.NotNil(t, rc.VirtualHosts[0].Routes[0])
			r := rc.VirtualHosts[0].Routes[0].GetRoute()
			require.NotNil(t, r)
			require.Len(t, r.UpgradeConfigs, 0)
		})
	})

	t.Run("With WebSocket", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
				UpgradeConfigs: ConfigDestinationsUpgrade{
					{
						Type: "websocket",
					},
				},
			},
		}, func(t *testing.T, b *bootstrapAPI.Bootstrap) {
			require.NotNil(t, b)
			require.NotNil(t, b.StaticResources)
			require.NotNil(t, b.StaticResources.Listeners)
			require.Len(t, b.StaticResources.Listeners, 1)
			require.NotNil(t, b.StaticResources.Listeners[0])
			require.NotNil(t, b.StaticResources.Listeners[0].DefaultFilterChain)
			require.NotNil(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters)
			require.Len(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters, 1)
			require.NotNil(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters[0])
			var o httpConnectionManagerAPI.HttpConnectionManager
			tgrpc.GRPCAnyCastAs(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters[0].GetTypedConfig(), &o)
			rc := o.GetRouteConfig()
			require.NotNil(t, rc)
			require.NotNil(t, rc.VirtualHosts)
			require.Len(t, rc.VirtualHosts, 1)
			require.NotNil(t, rc.VirtualHosts[0])
			require.Len(t, rc.VirtualHosts[0].Routes, 1)
			require.NotNil(t, rc.VirtualHosts[0].Routes[0])
			r := rc.VirtualHosts[0].Routes[0].GetRoute()
			require.NotNil(t, r)
			require.Len(t, r.UpgradeConfigs, 1)
			require.NotNil(t, r.UpgradeConfigs[0])
			require.EqualValues(t, "websocket", r.UpgradeConfigs[0].UpgradeType)
			require.NotNil(t, r.UpgradeConfigs[0].Enabled)
			require.True(t, r.UpgradeConfigs[0].Enabled.GetValue())
		})
	})

	t.Run("With Multi WebSocket", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
				UpgradeConfigs: ConfigDestinationsUpgrade{
					{
						Type: "websocket",
					},
					{
						Type:    "websocket",
						Enabled: util.NewType(false),
					},
				},
			},
		}, func(t *testing.T, b *bootstrapAPI.Bootstrap) {
			require.NotNil(t, b)
			require.NotNil(t, b.StaticResources)
			require.NotNil(t, b.StaticResources.Listeners)
			require.Len(t, b.StaticResources.Listeners, 1)
			require.NotNil(t, b.StaticResources.Listeners[0])
			require.NotNil(t, b.StaticResources.Listeners[0].DefaultFilterChain)
			require.NotNil(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters)
			require.Len(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters, 1)
			require.NotNil(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters[0])
			var o httpConnectionManagerAPI.HttpConnectionManager
			tgrpc.GRPCAnyCastAs(t, b.StaticResources.Listeners[0].DefaultFilterChain.Filters[0].GetTypedConfig(), &o)
			rc := o.GetRouteConfig()
			require.NotNil(t, rc)
			require.NotNil(t, rc.VirtualHosts)
			require.Len(t, rc.VirtualHosts, 1)
			require.NotNil(t, rc.VirtualHosts[0])
			require.Len(t, rc.VirtualHosts[0].Routes, 1)
			require.NotNil(t, rc.VirtualHosts[0].Routes[0])
			r := rc.VirtualHosts[0].Routes[0].GetRoute()
			require.NotNil(t, r)
			require.Len(t, r.UpgradeConfigs, 2)
			require.NotNil(t, r.UpgradeConfigs[0])
			require.NotNil(t, r.UpgradeConfigs[1])
			require.EqualValues(t, "websocket", r.UpgradeConfigs[0].UpgradeType)
			require.NotNil(t, r.UpgradeConfigs[0].Enabled)
			require.True(t, r.UpgradeConfigs[0].Enabled.GetValue())
			require.EqualValues(t, "websocket", r.UpgradeConfigs[1].UpgradeType)
			require.NotNil(t, r.UpgradeConfigs[1].Enabled)
			require.False(t, r.UpgradeConfigs[1].Enabled.GetValue())
		})
	})

	t.Run("Default", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
				Type: util.NewType(ConfigDestinationTypeHTTPS),
			},
			DefaultTLS: &ConfigTLS{
				CertificatePath: "/test",
				PrivateKeyPath:  "/test12",
			},
		})
	})

	t.Run("Default", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
				Path: util.NewType("/test/path/"),
				Type: util.NewType(ConfigDestinationTypeHTTPS),
			},
			DefaultTLS: &ConfigTLS{
				CertificatePath: "/test",
				PrivateKeyPath:  "/test12",
			},
		})
	})

	t.Run("Default", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
				Path: util.NewType("/test/path/"),
				Type: util.NewType(ConfigDestinationTypeHTTPS),
			},
			DefaultTLS: &ConfigTLS{
				CertificatePath: "/test",
				PrivateKeyPath:  "/test12",
			},
			Destinations: ConfigDestinations{
				"/test/": {
					Targets: []ConfigDestinationTarget{
						{
							Host: "127.0.0.1",
							Port: 12346,
						},
					},
					Path: util.NewType("/test/path/"),
					Type: util.NewType(ConfigDestinationTypeHTTPS),
				},
			},
		})
	})

	t.Run("Default", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
				Path: util.NewType("/test/path/"),
				Type: util.NewType(ConfigDestinationTypeHTTPS),
			},
			DefaultTLS: &ConfigTLS{
				CertificatePath: "/test",
				PrivateKeyPath:  "/test12",
			},
			Destinations: ConfigDestinations{
				"/_test/": {
					Targets: []ConfigDestinationTarget{
						{
							Host: "127.0.0.1",
							Port: 12346,
						},
					},
					Path: util.NewType("/test/path/"),
					Type: util.NewType(ConfigDestinationTypeHTTP),
				},
			},
		})
	})

	t.Run("Default", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
				Path: util.NewType("/test/path/"),
				Type: util.NewType(ConfigDestinationTypeHTTPS),
			},
			DefaultTLS: &ConfigTLS{
				CertificatePath: "/test",
				PrivateKeyPath:  "/test12",
			},
			SNI: []ConfigSNI{
				{
					ConfigTLS: ConfigTLS{
						CertificatePath: "/cp",
						PrivateKeyPath:  "/pp",
					},
					ServerNames: []string{
						"example.com",
					},
				},
			},
			Destinations: ConfigDestinations{
				"/_test/": {
					Targets: []ConfigDestinationTarget{
						{
							Host: "127.0.0.1",
							Port: 12346,
						},
					},
					Path: util.NewType("/test/path/"),
					Type: util.NewType(ConfigDestinationTypeHTTP),
				},
			},
		})
	})

	t.Run("Default", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
				Path: util.NewType("/test/path/"),
				Type: util.NewType(ConfigDestinationTypeHTTPS),
			},
			DefaultTLS: &ConfigTLS{
				CertificatePath: "/test",
				PrivateKeyPath:  "/test12",
			},
			SNI: []ConfigSNI{
				{
					ConfigTLS: ConfigTLS{
						CertificatePath: "/cp",
						PrivateKeyPath:  "/pp",
					},
					ServerNames: []string{
						"example.com",
					},
				},
				{
					ConfigTLS: ConfigTLS{
						CertificatePath: "/c2",
						PrivateKeyPath:  "/p2",
					},
					ServerNames: []string{
						"2.example.com",
					},
				},
			},
			Destinations: ConfigDestinations{
				"/_test/": {
					Targets: []ConfigDestinationTarget{
						{
							Host: "127.0.0.1",
							Port: 12346,
						},
					},
					Path: util.NewType("/test/path/"),
					Type: util.NewType(ConfigDestinationTypeHTTP),
				},
			},
		})
	})

	t.Run("Default", func(t *testing.T) {
		renderAndPrintGatewayConfig(t, Config{
			DefaultDestination: ConfigDestination{
				Targets: []ConfigDestinationTarget{
					{
						Host: "127.0.0.1",
						Port: 12345,
					},
				},
				Path: util.NewType("/test/path/"),
				Type: util.NewType(ConfigDestinationTypeHTTPS),
			},
			DefaultTLS: &ConfigTLS{
				CertificatePath: "/test",
				PrivateKeyPath:  "/test12",
			},
			SNI: []ConfigSNI{
				{
					ConfigTLS: ConfigTLS{
						CertificatePath: "/cp",
						PrivateKeyPath:  "/pp",
					},
					ServerNames: []string{
						"example.com",
					},
				},
				{
					ConfigTLS: ConfigTLS{
						CertificatePath: "/c2",
						PrivateKeyPath:  "/p2",
					},
					ServerNames: []string{
						"2.example.com",
					},
				},
			},
			Destinations: ConfigDestinations{
				"/_test/": {
					Targets: []ConfigDestinationTarget{
						{
							Host: "127.0.0.1",
							Port: 12346,
						},
					},
					Path: util.NewType("/test/path/"),
					Type: util.NewType(ConfigDestinationTypeHTTP),
				},
				"/_test2": {
					Type: util.NewType(ConfigDestinationTypeStatic),
					Static: &ConfigDestinationStatic{
						Code: util.NewType[uint32](302),
						Response: struct {
							Data string `json:"data"`
						}{
							Data: "TEST",
						},
					},
				},
			},
		})
	})
}
