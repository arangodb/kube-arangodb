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
	"time"

	bootstrapAPI "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterAPI "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	coreAPI "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointAPI "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerAPI "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routeAPI "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	routerAPI "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	httpConnectionManagerAPI "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsApi "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"sigs.k8s.io/yaml"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type GatewayConfig struct {
	DefaultAddress string `json:"defaultAddress,omitempty"`

	DefaultTLS *GatewayConfigTLS `json:"defaultTLS,omitempty"`
}

type GatewayConfigTLS struct {
	CertificatePath string `json:"certificatePath,omitempty"`
	PrivateKeyPath  string `json:"privateKeyPath,omitempty"`
}

func (g GatewayConfig) Validate() error {
	if g.DefaultAddress == "" {
		return errors.Errorf(".defaultAddress cannot be empty")
	}

	return nil
}

func (g GatewayConfig) RenderYAML() ([]byte, string, *bootstrapAPI.Bootstrap, error) {
	cfg, err := g.Render()
	if err != nil {
		return nil, "", nil, err
	}

	data, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(cfg)
	if err != nil {
		return nil, "", nil, err
	}

	data, err = yaml.JSONToYAML(data)
	return data, util.SHA256(data), cfg, err
}

func (g GatewayConfig) Render() (*bootstrapAPI.Bootstrap, error) {
	if err := g.Validate(); err != nil {
		return nil, errors.Wrapf(err, "Validation failed")
	}

	clusters, err := g.RenderClusters()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render clusters")
	}

	listener, err := g.RenderListener()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render listener")
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
				listener,
			},
			Clusters: clusters,
		},
	}, nil
}

func (g GatewayConfig) RenderClusters() ([]*clusterAPI.Cluster, error) {
	return []*clusterAPI.Cluster{
		{
			Name:           "default",
			ConnectTimeout: durationpb.New(time.Second),
			LbPolicy:       clusterAPI.Cluster_ROUND_ROBIN,
			LoadAssignment: &endpointAPI.ClusterLoadAssignment{
				ClusterName: "default",
				Endpoints: []*endpointAPI.LocalityLbEndpoints{
					{
						LbEndpoints: []*endpointAPI.LbEndpoint{
							{
								HostIdentifier: &endpointAPI.LbEndpoint_Endpoint{
									Endpoint: &endpointAPI.Endpoint{
										Address: &coreAPI.Address{
											Address: &coreAPI.Address_SocketAddress{
												SocketAddress: &coreAPI.SocketAddress{
													Address: g.DefaultAddress,
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
	}, nil
}

func (g GatewayConfig) RenderRoutes() ([]*routeAPI.Route, error) {
	return []*routeAPI.Route{
		{
			Match: &routeAPI.RouteMatch{
				PathSpecifier: &routeAPI.RouteMatch_Prefix{
					Prefix: "/",
				},
			},
			Action: &routeAPI.Route_Route{
				Route: &routeAPI.RouteAction{
					ClusterSpecifier: &routeAPI.RouteAction_Cluster{
						Cluster: "default",
					},
					PrefixRewrite: "/",
				},
			},
		},
	}, nil
}

func (g GatewayConfig) RenderFilters() ([]*listenerAPI.Filter, error) {
	httpFilterConfigType, err := anypb.New(&routerAPI.Router{})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render route config")
	}

	routes, err := g.RenderRoutes()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render routes")
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
		return nil, errors.Wrapf(err, "Unable to render http connection manager")
	}

	return []*listenerAPI.Filter{
		{
			Name: "envoy.filters.network.httpConnectionManagerAPI",
			ConfigType: &listenerAPI.Filter_TypedConfig{
				TypedConfig: filterConfigType,
			},
		},
	}, nil
}

func (g GatewayConfig) RenderDefaultFilterChain() (*listenerAPI.FilterChain, error) {
	filters, err := g.RenderFilters()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render filters")
	}

	ret := &listenerAPI.FilterChain{
		Filters: filters,
	}

	if tls := g.DefaultTLS; tls != nil {
		tlsContext, err := anypb.New(&tlsApi.DownstreamTlsContext{
			CommonTlsContext: &tlsApi.CommonTlsContext{
				TlsCertificates: []*tlsApi.TlsCertificate{
					{
						CertificateChain: &coreAPI.DataSource{
							Specifier: &coreAPI.DataSource_Filename{
								Filename: tls.CertificatePath,
							},
						},
						PrivateKey: &coreAPI.DataSource{
							Specifier: &coreAPI.DataSource_Filename{
								Filename: tls.PrivateKeyPath,
							},
						},
					},
				},
			},
		})
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to render tls context")
		}

		ret.TransportSocket = &coreAPI.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &coreAPI.TransportSocket_TypedConfig{
				TypedConfig: tlsContext,
			},
		}
	}

	return ret, nil
}

func (g GatewayConfig) RenderSecondaryFilterChains() ([]*listenerAPI.FilterChain, error) {
	return nil, nil
}

func (g GatewayConfig) RenderListener() (*listenerAPI.Listener, error) {
	filterChains, err := g.RenderSecondaryFilterChains()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render secondary filter chains")
	}

	defaultFilterChain, err := g.RenderDefaultFilterChain()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render default filter")
	}

	return &listenerAPI.Listener{
		Name: "default",
		Address: &coreAPI.Address{
			Address: &coreAPI.Address_SocketAddress{
				SocketAddress: &coreAPI.SocketAddress{
					Address:       "0.0.0.0",
					PortSpecifier: &coreAPI.SocketAddress_PortValue{PortValue: shared.ArangoPort},
				},
			},
		},
		FilterChains: filterChains,

		DefaultFilterChain: defaultFilterChain,
	}, nil
}
