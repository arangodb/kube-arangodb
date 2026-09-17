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

package gateway

import (
	"time"

	pbEnvoyBootstrapV3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	pbEnvoyClusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyEndpointV3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	upstreamHttpApi "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/arangodb/kube-arangodb/pkg/util"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

// NodeADSConfig renders the Envoy bootstrap for push (ADS) mode: a static cluster pointing at the
// integration sidecar's ADS server, and dynamic_resources that subscribe to CDS and LDS over that ADS
// stream. Unlike the filesystem mode, the config is delivered by the sidecar over gRPC, so changes
// propagate without waiting for the kubelet ConfigMap sync.
func NodeADSConfig(cluster, id, adsClusterName string, target ConfigDestinationTarget) ([]byte, string, *pbEnvoyBootstrapV3.Bootstrap, error) {
	adsCluster, err := renderHTTP2UnixCluster(adsClusterName, target)
	if err != nil {
		return nil, "", nil, err
	}

	b := &pbEnvoyBootstrapV3.Bootstrap{
		Node: &pbEnvoyCoreV3.Node{
			Id:      id,
			Cluster: cluster,
		},
		StaticResources: &pbEnvoyBootstrapV3.Bootstrap_StaticResources{
			Clusters: []*pbEnvoyClusterV3.Cluster{
				adsCluster,
			},
		},
		DynamicResources: &pbEnvoyBootstrapV3.Bootstrap_DynamicResources{
			AdsConfig: &pbEnvoyCoreV3.ApiConfigSource{
				ApiType:             pbEnvoyCoreV3.ApiConfigSource_GRPC,
				TransportApiVersion: pbEnvoyCoreV3.ApiVersion_V3,
				GrpcServices: []*pbEnvoyCoreV3.GrpcService{
					{
						TargetSpecifier: &pbEnvoyCoreV3.GrpcService_EnvoyGrpc_{
							EnvoyGrpc: &pbEnvoyCoreV3.GrpcService_EnvoyGrpc{
								ClusterName: adsClusterName,
							},
						},
					},
				},
			},
			CdsConfig: adsConfigSource(),
			LdsConfig: adsConfigSource(),
		},
	}

	data, err := ugrpc.MarshalYAML(b, ugrpc.WithUseProtoNames(true))
	if err != nil {
		return nil, "", nil, err
	}

	return data, util.SHA256(data), b, nil
}

// adsConfigSource returns a ConfigSource that resolves via the aggregated (ADS) stream.
func adsConfigSource() *pbEnvoyCoreV3.ConfigSource {
	return &pbEnvoyCoreV3.ConfigSource{
		ResourceApiVersion: pbEnvoyCoreV3.ApiVersion_V3,
		ConfigSourceSpecifier: &pbEnvoyCoreV3.ConfigSource_Ads{
			Ads: &pbEnvoyCoreV3.AggregatedConfigSource{},
		},
	}
}

// renderHTTP2UnixCluster builds a STATIC HTTP/2 cluster targeting the given (unix socket) endpoint,
// suitable for gRPC/ADS. It mirrors the integration sidecar cluster rendered into CDS.
func renderHTTP2UnixCluster(name string, target ConfigDestinationTarget) (*pbEnvoyClusterV3.Cluster, error) {
	hpo, err := anypb.New(&upstreamHttpApi.HttpProtocolOptions{
		UpstreamProtocolOptions: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
					Http2ProtocolOptions: &pbEnvoyCoreV3.Http2ProtocolOptions{
						ConnectionKeepalive: &pbEnvoyCoreV3.KeepaliveSettings{
							Interval:               durationpb.New(15 * time.Second),
							Timeout:                durationpb.New(30 * time.Second),
							ConnectionIdleInterval: durationpb.New(60 * time.Second),
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &pbEnvoyClusterV3.Cluster{
		Name:                 name,
		ConnectTimeout:       durationpb.New(time.Second),
		LbPolicy:             pbEnvoyClusterV3.Cluster_ROUND_ROBIN,
		ClusterDiscoveryType: evaluateClusterDiscoveryType(target),
		LoadAssignment: &pbEnvoyEndpointV3.ClusterLoadAssignment{
			ClusterName: name,
			Endpoints: []*pbEnvoyEndpointV3.LocalityLbEndpoints{
				{
					LbEndpoints: []*pbEnvoyEndpointV3.LbEndpoint{
						target.RenderEndpoint(),
					},
				},
			},
		},
		TypedExtensionProtocolOptions: map[string]*anypb.Any{
			"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": hpo,
		},
	}, nil
}
