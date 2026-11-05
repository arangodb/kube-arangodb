//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
	"testing"

	tlsApi "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_GatewaySDS(t *testing.T) {
	cfg := Config{
		DefaultDestination: ConfigDestination{
			Targets: []ConfigDestinationTarget{
				ConfigDestinationTargetEndpoint{Host: "127.0.0.1", Port: 8529},
			},
			Type: util.NewType(ConfigDestinationTypeHTTPS),
		},
		DefaultTLS: &ConfigTLS{
			Name:            "internal",
			CertificatePath: "/secrets/tls/tls.keyfile",
			PrivateKeyPath:  "/secrets/tls/tls.keyfile",
			WatchDir:        "/secrets/tls",
			SDSPath:         "/etc/gateway/sds/internal.yaml",
		},
		SNI: ConfigSNIList{
			{
				ConfigTLS: ConfigTLS{
					Name:            "sni-example",
					CertificatePath: "/secrets/sni/example/tls.keyfile",
					PrivateKeyPath:  "/secrets/sni/example/tls.keyfile",
					WatchDir:        "/secrets/sni/example",
					SDSPath:         "/etc/gateway/sds/sni-example.yaml",
				},
				ServerNames: []string{"example.com"},
			},
		},
	}

	t.Run("RenderSDS renders one file per certificate", func(t *testing.T) {
		files, err := cfg.RenderSDS()
		require.NoError(t, err)
		require.Len(t, files, 2)

		require.Contains(t, files, "internal.yaml")
		require.Contains(t, files, "sni-example.yaml")

		// Each SDS definition references the mounted keyfile by path (stable across rotations).
		require.Contains(t, files["internal.yaml"], "/secrets/tls/tls.keyfile")
		require.Contains(t, files["sni-example.yaml"], "/secrets/sni/example/tls.keyfile")
	})

	// The listener must deliver certificates via SDS (not inline) and watch the mounted secret
	// directory, so a rotated keyfile is reloaded in place without a restart.
	assertSDS := func(t *testing.T, tls *ConfigTLS) {
		t.Helper()

		ts, err := tls.RenderListenerTransportSocket()
		require.NoError(t, err)
		require.NotNil(t, ts)

		var dtc tlsApi.DownstreamTlsContext
		require.NoError(t, ts.GetTypedConfig().UnmarshalTo(&dtc))

		require.Empty(t, dtc.GetCommonTlsContext().GetTlsCertificates(), "certificate must not be inlined")

		sds := dtc.GetCommonTlsContext().GetTlsCertificateSdsSecretConfigs()
		require.Len(t, sds, 1)
		require.Equal(t, tls.Name, sds[0].GetName())

		pcs := sds[0].GetSdsConfig().GetPathConfigSource()
		require.NotNil(t, pcs)
		require.Equal(t, tls.SDSPath, pcs.GetPath())
		require.Equal(t, tls.WatchDir, pcs.GetWatchedDirectory().GetPath())
	}

	t.Run("Internal listener transport socket uses SDS", func(t *testing.T) {
		assertSDS(t, cfg.DefaultTLS)
	})

	t.Run("SNI listener transport socket uses SDS", func(t *testing.T) {
		assertSDS(t, &cfg.SNI[0].ConfigTLS)
	})
}
