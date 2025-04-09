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
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyEndpointV3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigDestinationTargets []ConfigDestinationTarget

func (c ConfigDestinationTargets) RenderEndpoints() []*pbEnvoyEndpointV3.LbEndpoint {
	var endpoints = make([]*pbEnvoyEndpointV3.LbEndpoint, len(c))

	for id := range c {
		endpoints[id] = c[id].RenderEndpoint()
	}

	return endpoints
}

func (c ConfigDestinationTargets) Validate() error {
	if len(c) == 0 {
		return errors.Errorf("Empty Target not allowed")
	}
	return shared.ValidateList(c, func(target ConfigDestinationTarget) error {
		return target.Validate()
	})
}

type ConfigDestinationTarget struct {
	Host string `json:"ip,omitempty"`
	Port int32  `json:"port,omitempty"`
}

func (c *ConfigDestinationTarget) Validate() error {
	if c == nil {
		return nil
	}
	return shared.WithErrors(
		shared.ValidateRequiredPath("ip", &c.Host, func(t string) error {
			if t == "" {
				return errors.Errorf("Empty string not allowed")
			}
			return nil
		}),
		shared.ValidateRequiredPath("ip", &c.Port, func(t int32) error {
			if t <= 0 {
				return errors.Errorf("Port needs to be greater than 0")
			}
			return nil
		}),
	)
}

func (c *ConfigDestinationTarget) RenderEndpoint() *pbEnvoyEndpointV3.LbEndpoint {
	if c == nil {
		return nil
	}
	return &pbEnvoyEndpointV3.LbEndpoint{
		HostIdentifier: &pbEnvoyEndpointV3.LbEndpoint_Endpoint{
			Endpoint: &pbEnvoyEndpointV3.Endpoint{
				Address: &pbEnvoyCoreV3.Address{
					Address: &pbEnvoyCoreV3.Address_SocketAddress{
						SocketAddress: &pbEnvoyCoreV3.SocketAddress{
							Protocol: pbEnvoyCoreV3.SocketAddress_TCP,
							Address:  c.Host,
							PortSpecifier: &pbEnvoyCoreV3.SocketAddress_PortValue{
								PortValue: uint32(c.Port),
							},
						},
					},
				},
			},
		},
	}
}
