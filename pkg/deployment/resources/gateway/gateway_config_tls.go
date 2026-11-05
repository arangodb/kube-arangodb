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
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsApi "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigTLS struct {
	// Name is the unique SDS secret name used to reference this certificate from the listener.
	Name string `json:"name,omitempty"`

	CertificatePath string `json:"certificatePath,omitempty"`
	PrivateKeyPath  string `json:"privateKeyPath,omitempty"`

	// WatchDir is the directory Envoy watches for atomic changes - the mounted Kubernetes secret
	// directory. When the secret rotates, Envoy reloads the SDS secret in place (no pod restart).
	WatchDir string `json:"watchDir,omitempty"`

	// SDSPath is the filesystem path of the SDS secret definition file that Envoy sources.
	SDSPath string `json:"sdsPath,omitempty"`
}

// RenderListenerTransportSocket renders the downstream TLS transport socket. The certificate is
// delivered via filesystem SDS (Secret Discovery Service) rather than inlined so that Envoy can
// hot-reload a rotated certificate in place, without churning the listener or restarting the pod.
func (c *ConfigTLS) RenderListenerTransportSocket() (*pbEnvoyCoreV3.TransportSocket, error) {
	if c == nil {
		return nil, nil
	}

	tlsContext, err := anypb.New(&tlsApi.DownstreamTlsContext{
		CommonTlsContext: &tlsApi.CommonTlsContext{
			TlsCertificateSdsSecretConfigs: []*tlsApi.SdsSecretConfig{
				{
					Name:      c.Name,
					SdsConfig: c.renderSDSConfigSource(),
				},
			},
			AlpnProtocols: []string{(ALPNProtocolHTTP2 | ALPNProtocolHTTP1).String()},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render tls context")
	}

	return &pbEnvoyCoreV3.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &pbEnvoyCoreV3.TransportSocket_TypedConfig{
			TypedConfig: tlsContext,
		},
	}, nil
}

// renderSDSConfigSource builds a filesystem SDS config source: Envoy sources the secret definition
// from SDSPath and reloads it whenever a file is moved into WatchDir (the mounted Kubernetes secret
// directory, which is updated by an atomic symlink swap on rotation). This lets a rotated
// certificate be picked up in place without restarting the gateway.
func (c *ConfigTLS) renderSDSConfigSource() *pbEnvoyCoreV3.ConfigSource {
	src := &pbEnvoyCoreV3.PathConfigSource{
		Path: c.SDSPath,
	}
	if c.WatchDir != "" {
		src.WatchedDirectory = &pbEnvoyCoreV3.WatchedDirectory{
			Path: c.WatchDir,
		}
	}

	return &pbEnvoyCoreV3.ConfigSource{
		ResourceApiVersion: pbEnvoyCoreV3.ApiVersion_V3,
		ConfigSourceSpecifier: &pbEnvoyCoreV3.ConfigSource_PathConfigSource{
			PathConfigSource: src,
		},
	}
}

// RenderSDSSecret builds the Envoy SDS Secret that points at the mounted certificate keyfile. The
// referenced file path is stable across rotations; Envoy re-reads its contents when WatchDir changes.
func (c *ConfigTLS) RenderSDSSecret() *tlsApi.Secret {
	if c == nil {
		return nil
	}

	return &tlsApi.Secret{
		Name: c.Name,
		Type: &tlsApi.Secret_TlsCertificate{
			TlsCertificate: &tlsApi.TlsCertificate{
				CertificateChain: &pbEnvoyCoreV3.DataSource{
					Specifier: &pbEnvoyCoreV3.DataSource_Filename{
						Filename: c.CertificatePath,
					},
				},
				PrivateKey: &pbEnvoyCoreV3.DataSource{
					Specifier: &pbEnvoyCoreV3.DataSource_Filename{
						Filename: c.PrivateKeyPath,
					},
				},
			},
		},
	}
}
