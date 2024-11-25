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
	"time"

	coreAPI "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	upstreamHttpApi "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigDestinationProtocol int

const (
	ConfigDestinationProtocolHTTP1 ConfigDestinationProtocol = iota
	ConfigDestinationProtocolHTTP2
)

func (c *ConfigDestinationProtocol) Get() ConfigDestinationProtocol {
	if c == nil {
		return ConfigDestinationProtocolHTTP1
	}

	switch v := *c; v {
	case ConfigDestinationProtocolHTTP1, ConfigDestinationProtocolHTTP2:
		return v
	default:
		return ConfigDestinationProtocolHTTP1
	}
}

func (c *ConfigDestinationProtocol) Options() *upstreamHttpApi.HttpProtocolOptions {
	switch c.Get() {
	case ConfigDestinationProtocolHTTP1:
		return &upstreamHttpApi.HttpProtocolOptions{
			UpstreamProtocolOptions: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{
						HttpProtocolOptions: &coreAPI.Http1ProtocolOptions{},
					},
				},
			},
		}
	case ConfigDestinationProtocolHTTP2:
		return &upstreamHttpApi.HttpProtocolOptions{
			UpstreamProtocolOptions: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
						Http2ProtocolOptions: &coreAPI.Http2ProtocolOptions{
							ConnectionKeepalive: &coreAPI.KeepaliveSettings{
								Interval:               durationpb.New(15 * time.Second),
								Timeout:                durationpb.New(30 * time.Second),
								ConnectionIdleInterval: durationpb.New(60 * time.Second),
							},
						},
					},
				},
			},
		}
	default:
		return &upstreamHttpApi.HttpProtocolOptions{
			UpstreamProtocolOptions: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{
						HttpProtocolOptions: &coreAPI.Http1ProtocolOptions{},
					},
				},
			},
		}
	}
}

func (c *ConfigDestinationProtocol) Validate() error {
	switch c.Get() {
	case ConfigDestinationProtocolHTTP1, ConfigDestinationProtocolHTTP2:
		return nil
	default:
		return errors.Errorf("Invalid destination protocol")
	}
}
