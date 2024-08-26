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

package resources

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	bootstrapAPI "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterAPI "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	coreAPI "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointAPI "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerAPI "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routeAPI "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	routerAPI "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	httpConnectionManagerAPI "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"sigs.k8s.io/yaml"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Redirect util.KV[string, []string]

func WithRedirect(prefix string, target ...string) Redirect {
	return Redirect{
		K: prefix,
		V: target,
	}
}

func RenderGatewayConfigYAML(dbServiceAddress string, redirects ...Redirect) ([]byte, error) {
	cfg, err := RenderConfig(dbServiceAddress, redirects...)
	if err != nil {
		return nil, err
	}

	data, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	data, err = yaml.JSONToYAML(data)
	return data, err
}

func RenderConfig(dbServiceAddress string, redirects ...Redirect) (*bootstrapAPI.Bootstrap, error) {
	clusters := []*clusterAPI.Cluster{
		{
			Name:           "arangodb",
			ConnectTimeout: durationpb.New(250 * time.Millisecond),
			LbPolicy:       clusterAPI.Cluster_ROUND_ROBIN,
			LoadAssignment: &endpointAPI.ClusterLoadAssignment{
				ClusterName: "arangodb",
				Endpoints: []*endpointAPI.LocalityLbEndpoints{
					{
						LbEndpoints: []*endpointAPI.LbEndpoint{
							{
								HostIdentifier: &endpointAPI.LbEndpoint_Endpoint{
									Endpoint: &endpointAPI.Endpoint{
										Address: &coreAPI.Address{
											Address: &coreAPI.Address_SocketAddress{
												SocketAddress: &coreAPI.SocketAddress{
													Address: dbServiceAddress,
													PortSpecifier: &coreAPI.SocketAddress_PortValue{
														PortValue: shared.ArangoPort,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	routes := []*routeAPI.Route{
		{
			Match: &routeAPI.RouteMatch{
				PathSpecifier: &routeAPI.RouteMatch_Prefix{
					Prefix: "/",
				},
			},
			Action: &routeAPI.Route_Route{
				Route: &routeAPI.RouteAction{
					ClusterSpecifier: &routeAPI.RouteAction_Cluster{
						Cluster: "arangodb",
					},
					PrefixRewrite: "/",
				},
			},
		},
	}

	for id, redirect := range redirects {
		var endpoints []*endpointAPI.LbEndpoint

		for _, target := range redirect.V {
			req, err := url.Parse(target)
			if err != nil {
				return nil, err
			}

			port, err := strconv.Atoi(req.Port())
			if err != nil {
				return nil, err
			}

			endpoints = append(endpoints, &endpointAPI.LbEndpoint{
				HostIdentifier: &endpointAPI.LbEndpoint_Endpoint{
					Endpoint: &endpointAPI.Endpoint{
						Address: &coreAPI.Address{
							Address: &coreAPI.Address_SocketAddress{
								SocketAddress: &coreAPI.SocketAddress{
									Address: req.Hostname(),
									PortSpecifier: &coreAPI.SocketAddress_PortValue{
										PortValue: uint32(port),
									},
								},
							},
						},
					},
				},
			},
			)
		}

		cluster := &clusterAPI.Cluster{
			Name:           fmt.Sprintf("cluster_%05d", id),
			ConnectTimeout: durationpb.New(250 * time.Millisecond),
			LbPolicy:       clusterAPI.Cluster_ROUND_ROBIN,
			LoadAssignment: &endpointAPI.ClusterLoadAssignment{
				ClusterName: fmt.Sprintf("cluster_%05d", id),
				Endpoints: []*endpointAPI.LocalityLbEndpoints{
					{
						LbEndpoints: endpoints,
					},
				},
			},
		}

		route := &routeAPI.Route{
			Match: &routeAPI.RouteMatch{
				PathSpecifier: &routeAPI.RouteMatch_Prefix{
					Prefix: redirect.K,
				},
			},
			Action: &routeAPI.Route_Route{
				Route: &routeAPI.RouteAction{
					ClusterSpecifier: &routeAPI.RouteAction_Cluster{
						Cluster: fmt.Sprintf("cluster_%05d", id),
					},
					PrefixRewrite: "/",
				},
			},
		}

		clusters = append(clusters, cluster)
		routes = append(routes, route)
	}

	routes = util.Sort(routes, func(i, j *routeAPI.Route) bool {
		return i.Match.GetPrefix() > j.Match.GetPrefix()
	})

	httpFilterConfigType, err := anypb.New(&routerAPI.Router{})
	if err != nil {
		return nil, err
	}

	filterConfigType, err := anypb.New(&httpConnectionManagerAPI.HttpConnectionManager{
		StatPrefix: "ingress_http",
		CodecType:  httpConnectionManagerAPI.HttpConnectionManager_AUTO,
		RouteSpecifier: &httpConnectionManagerAPI.HttpConnectionManager_RouteConfig{
			RouteConfig: &routeAPI.RouteConfiguration{
				Name: "local_route",
				VirtualHosts: []*routeAPI.VirtualHost{
					{
						Name:    "local_service",
						Domains: []string{"*"},
						Routes:  routes,
					},
				},
			},
		},
		HttpFilters: []*httpConnectionManagerAPI.HttpFilter{
			{
				Name: "envoy.filters.http.routerAPI",
				ConfigType: &httpConnectionManagerAPI.HttpFilter_TypedConfig{
					TypedConfig: httpFilterConfigType,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &bootstrapAPI.Bootstrap{
		Admin: &bootstrapAPI.Admin{
			Address: &coreAPI.Address{
				Address: &coreAPI.Address_SocketAddress{
					SocketAddress: &coreAPI.SocketAddress{
						Address:       "127.0.0.1",
						PortSpecifier: &coreAPI.SocketAddress_PortValue{PortValue: 9901},
					},
				},
			},
		},
		StaticResources: &bootstrapAPI.Bootstrap_StaticResources{
			Listeners: []*listenerAPI.Listener{
				{
					Name: "listener_0",
					Address: &coreAPI.Address{
						Address: &coreAPI.Address_SocketAddress{
							SocketAddress: &coreAPI.SocketAddress{
								Address:       "0.0.0.0",
								PortSpecifier: &coreAPI.SocketAddress_PortValue{PortValue: shared.ArangoPort},
							},
						},
					},
					FilterChains: []*listenerAPI.FilterChain{
						{
							Filters: []*listenerAPI.Filter{
								{
									Name: "envoy.filters.network.httpConnectionManagerAPI",
									ConfigType: &listenerAPI.Filter_TypedConfig{
										TypedConfig: filterConfigType,
									},
								},
							},
						},
					},
				},
			},
			Clusters: clusters,
		},
	}, nil
}
