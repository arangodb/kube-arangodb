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
	coreAPI "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsApi "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigDestinationType int

const (
	ConfigDestinationTypeHTTP ConfigDestinationType = iota
	ConfigDestinationTypeHTTPS
	ConfigDestinationTypeStatic
)

func (c *ConfigDestinationType) Get() ConfigDestinationType {
	if c == nil {
		return ConfigDestinationTypeHTTP
	}

	switch v := *c; v {
	case ConfigDestinationTypeHTTP, ConfigDestinationTypeHTTPS, ConfigDestinationTypeStatic:
		return v
	default:
		return ConfigDestinationTypeHTTP
	}
}

func (c *ConfigDestinationType) RenderUpstreamTransportSocket(protocol *ConfigDestinationProtocol, config ConfigDestinationTLS) (*coreAPI.TransportSocket, error) {
	if c.Get() == ConfigDestinationTypeHTTPS {
		tlsConfig, err := anypb.New(&tlsApi.UpstreamTlsContext{
			CommonTlsContext: &tlsApi.CommonTlsContext{
				AlpnProtocols: []string{protocol.ALPN().String()},
				ValidationContextType: &tlsApi.CommonTlsContext_ValidationContext{
					ValidationContext: &tlsApi.CertificateValidationContext{
						TrustChainVerification: util.BoolSwitch(!config.IsInsecure(), tlsApi.CertificateValidationContext_VERIFY_TRUST_CHAIN, tlsApi.CertificateValidationContext_ACCEPT_UNTRUSTED),
					},
				},
			},
		})
		if err != nil {
			return nil, err
		}

		return &coreAPI.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &coreAPI.TransportSocket_TypedConfig{
				TypedConfig: tlsConfig,
			},
		}, nil
	}

	return nil, nil
}

func (c *ConfigDestinationType) Validate() error {
	switch c.Get() {
	case ConfigDestinationTypeHTTP, ConfigDestinationTypeHTTPS, ConfigDestinationTypeStatic:
		return nil
	default:
		return errors.Errorf("Invalid destination type")
	}
}
