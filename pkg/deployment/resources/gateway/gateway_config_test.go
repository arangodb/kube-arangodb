//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	bootstrapAPI "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func renderAndPrintGatewayConfig(t *testing.T, cfg Config) *bootstrapAPI.Bootstrap {
	data, checksum, obj, err := cfg.RenderYAML()
	require.NoError(t, err)

	t.Logf("Checksum: %s", checksum)
	t.Log(string(data))

	return obj
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
}
