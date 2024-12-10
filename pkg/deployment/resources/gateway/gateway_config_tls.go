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
	coreAPI "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsApi "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigTLS struct {
	CertificatePath string `json:"certificatePath,omitempty"`
	PrivateKeyPath  string `json:"privateKeyPath,omitempty"`
}

func (c *ConfigTLS) RenderListenerTransportSocket() (*coreAPI.TransportSocket, error) {
	if c == nil {
		return nil, nil
	}

	tlsContext, err := anypb.New(&tlsApi.DownstreamTlsContext{
		CommonTlsContext: &tlsApi.CommonTlsContext{
			TlsCertificates: []*tlsApi.TlsCertificate{
				{
					CertificateChain: &coreAPI.DataSource{
						Specifier: &coreAPI.DataSource_Filename{
							Filename: c.CertificatePath,
						},
					},
					PrivateKey: &coreAPI.DataSource{
						Specifier: &coreAPI.DataSource_Filename{
							Filename: c.PrivateKeyPath,
						},
					},
				},
			},
			AlpnProtocols: []string{(ALPNProtocolHTTP2 | ALPNProtocolHTTP1).String()},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render tls context")
	}

	return &coreAPI.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &coreAPI.TransportSocket_TypedConfig{
			TypedConfig: tlsContext,
		},
	}, nil
}
