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
	"fmt"
	"sort"
	"time"

	pbEnvoyBootstrapV3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	pbEnvoyClusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyEndpointV3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	pbEnvoyListenerV3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	pbEnvoyRouteV3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	httpFilterAuthzApi "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	routerAPI "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	tlsInspectorApi "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	httpConnectionManagerAPI "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	upstreamHttpApi "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	discoveryApi "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

type Config struct {
	DefaultDestination ConfigDestination `json:"defaultDestination,omitempty"`

	Destinations ConfigDestinations `json:"destinations,omitempty"`

	DefaultTLS *ConfigTLS `json:"defaultTLS,omitempty"`

	IntegrationSidecar *ConfigDestinationTarget `json:"integrationSidecar,omitempty"`

	SNI ConfigSNIList `json:"sni,omitempty"`

	Options *ConfigOptions `json:"options,omitempty"`
}

func (c Config) Validate() error {
	return errors.Errors(
		shared.PrefixResourceErrors("defaultDestination", c.DefaultDestination.Validate()),
		shared.PrefixResourceErrors("integrationSidecar", c.IntegrationSidecar.Validate()),
		shared.PrefixResourceErrors("destinations", c.Destinations.Validate()),
		shared.PrefixResourceErrors("sni", c.SNI.Validate()),
	)
}

func (c Config) RenderYAML() ([]byte, string, *pbEnvoyBootstrapV3.Bootstrap, error) {
	cfg, err := c.Render()
	if err != nil {
		return nil, "", nil, err
	}

	data, err := ugrpc.MarshalYAML(cfg, ugrpc.WithUseProtoNames(true))
	if err != nil {
		return nil, "", nil, err
	}
	return data, util.SHA256(data), cfg, nil
}

func (c Config) RenderCDSYAML() ([]byte, string, *discoveryApi.DiscoveryResponse, error) {
	cfg, err := c.RenderCDS()
	if err != nil {
		return nil, "", nil, err
	}

	data, err := ugrpc.MarshalYAML(cfg, ugrpc.WithUseProtoNames(true))
	if err != nil {
		return nil, "", nil, err
	}
	return data, util.SHA256(data), cfg, nil
}

func (c Config) RenderLDSYAML() ([]byte, string, *discoveryApi.DiscoveryResponse, error) {
	cfg, err := c.RenderLDS()
	if err != nil {
		return nil, "", nil, err
	}

	data, err := ugrpc.MarshalYAML(cfg, ugrpc.WithUseProtoNames(true))
	if err != nil {
		return nil, "", nil, err
	}
	return data, util.SHA256(data), cfg, nil
}

func (c Config) RenderCDS() (*discoveryApi.DiscoveryResponse, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.Wrapf(err, "Validation failed")
	}

	clusters, err := c.RenderClusters()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render clusters")
	}

	return DynamicConfigResponse(clusters...)
}

func (c Config) RenderLDS() (*discoveryApi.DiscoveryResponse, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.Wrapf(err, "Validation failed")
	}

	listener, err := c.RenderListener()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render listener")
	}

	return DynamicConfigResponse(listener)
}

func (c Config) Render() (*pbEnvoyBootstrapV3.Bootstrap, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.Wrapf(err, "Validation failed")
	}

	clusters, err := c.RenderClusters()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render clusters")
	}

	listener, err := c.RenderListener()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render listener")
	}

	return &pbEnvoyBootstrapV3.Bootstrap{
		Admin: &pbEnvoyBootstrapV3.Admin{
			Address: &pbEnvoyCoreV3.Address{
				Address: &pbEnvoyCoreV3.Address_SocketAddress{
					SocketAddress: &pbEnvoyCoreV3.SocketAddress{
						Address:       "127.0.0.1",
						PortSpecifier: &pbEnvoyCoreV3.SocketAddress_PortValue{PortValue: 9901},
					},
				},
			},
		},
		StaticResources: &pbEnvoyBootstrapV3.Bootstrap_StaticResources{
			Listeners: []*pbEnvoyListenerV3.Listener{
				listener,
			},
			Clusters: clusters,
		},
	}, nil
}

func (c Config) RenderClusters() ([]*pbEnvoyClusterV3.Cluster, error) {
	def, err := c.DefaultDestination.RenderCluster("default")
	if err != nil {
		return nil, err
	}
	clusters := []*pbEnvoyClusterV3.Cluster{
		def,
	}

	if i := c.IntegrationSidecar; i != nil {
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
		cluster := &pbEnvoyClusterV3.Cluster{
			Name:           utilConstants.EnvoyIntegrationSidecarCluster,
			ConnectTimeout: durationpb.New(time.Second),
			LbPolicy:       pbEnvoyClusterV3.Cluster_ROUND_ROBIN,
			LoadAssignment: &pbEnvoyEndpointV3.ClusterLoadAssignment{
				ClusterName: utilConstants.EnvoyIntegrationSidecarCluster,
				Endpoints: []*pbEnvoyEndpointV3.LocalityLbEndpoints{
					{
						LbEndpoints: []*pbEnvoyEndpointV3.LbEndpoint{
							i.RenderEndpoint(),
						},
					},
				},
			},
			TypedExtensionProtocolOptions: map[string]*anypb.Any{
				"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": hpo,
			},
		}

		clusters = append(clusters, cluster)
	}

	for k, v := range c.Destinations {
		name := fmt.Sprintf("cluster_%s", util.SHA256FromString(k))
		c, err := v.RenderCluster(name)
		if err != nil {
			return nil, err
		}

		if c == nil {
			continue
		}
		clusters = append(clusters, c)
	}

	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Name < clusters[j].Name
	})

	return clusters, nil
}

func (c Config) RenderRoutes() ([]*pbEnvoyRouteV3.Route, error) {
	def, err := c.DefaultDestination.RenderRoute("default", "/")
	if err != nil {
		return nil, err
	}
	routes := []*pbEnvoyRouteV3.Route{
		def,
	}

	for k, v := range c.Destinations {
		name := fmt.Sprintf("cluster_%s", util.SHA256FromString(k))
		c, err := v.RenderRoute(name, k)
		if err != nil {
			return nil, err
		}

		routes = append(routes, c)
	}

	sort.SliceStable(routes, func(i, j int) bool {
		iPath := routes[i].GetMatch().GetPath()
		iPrefix := routes[i].GetMatch().GetPrefix()

		jPath := routes[j].GetMatch().GetPath()
		jPrefix := routes[j].GetMatch().GetPrefix()

		if iPath != "" && jPath != "" {
			return iPath > jPath
		} else if iPath == "" && jPath != "" {
			return false
		} else if iPath != "" && jPath == "" {
			return true
		} else {
			return iPrefix > jPrefix
		}
	})

	return routes, nil
}

func (c Config) RenderIntegrationSidecarFilter() (*httpConnectionManagerAPI.HttpFilter, error) {
	e, err := anypb.New(&httpFilterAuthzApi.ExtAuthz{
		Services: &httpFilterAuthzApi.ExtAuthz_GrpcService{
			GrpcService: &pbEnvoyCoreV3.GrpcService{
				TargetSpecifier: &pbEnvoyCoreV3.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &pbEnvoyCoreV3.GrpcService_EnvoyGrpc{
						ClusterName: "integration_sidecar",
					},
				},
				Timeout: durationpb.New(2 * time.Second),
			},
		},
		IncludePeerCertificate: true,
	})
	if err != nil {
		return nil, err
	}

	return &httpConnectionManagerAPI.HttpFilter{
		Name: utilConstants.EnvoyIntegrationSidecarFilterName,
		ConfigType: &httpConnectionManagerAPI.HttpFilter_TypedConfig{
			TypedConfig: e,
		},
		IsOptional: false,
	}, nil
}

func (c Config) RenderFilters() ([]*pbEnvoyListenerV3.Filter, error) {
	httpFilterConfigType, err := anypb.New(&routerAPI.Router{})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render route config")
	}

	routes, err := c.RenderRoutes()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render routes")
	}

	var httpFilters []*httpConnectionManagerAPI.HttpFilter

	if i := c.IntegrationSidecar; i != nil {
		q, err := c.RenderIntegrationSidecarFilter()
		if err != nil {
			return nil, err
		}
		httpFilters = append(httpFilters, q)
	}

	filterConfigType, err := anypb.New(&httpConnectionManagerAPI.HttpConnectionManager{
		StatPrefix:                 "ingress_http",
		CodecType:                  httpConnectionManagerAPI.HttpConnectionManager_AUTO,
		ServerHeaderTransformation: httpConnectionManagerAPI.HttpConnectionManager_PASS_THROUGH,
		MergeSlashes:               c.Options.GetMergeSlashes(),

		RouteSpecifier: &httpConnectionManagerAPI.HttpConnectionManager_RouteConfig{
			RouteConfig: &pbEnvoyRouteV3.RouteConfiguration{
				Name:                           "default",
				MaxDirectResponseBodySizeBytes: wrapperspb.UInt32(utilConstants.MaxInventorySize),
				VirtualHosts: []*pbEnvoyRouteV3.VirtualHost{
					{
						Name:    "default",
						Domains: []string{"*"},
						Routes:  routes,
					},
				},
				ValidateClusters: &wrapperspb.BoolValue{
					Value: false,
				},
			},
		},
		HttpFilters: append(httpFilters, &httpConnectionManagerAPI.HttpFilter{
			Name: "envoy.filters.http.routerAPI",
			ConfigType: &httpConnectionManagerAPI.HttpFilter_TypedConfig{
				TypedConfig: httpFilterConfigType,
			},
		},
		),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render http connection manager")
	}

	return []*pbEnvoyListenerV3.Filter{
		{
			Name: "envoy.filters.network.httpConnectionManagerAPI",
			ConfigType: &pbEnvoyListenerV3.Filter_TypedConfig{
				TypedConfig: filterConfigType,
			},
		},
	}, nil
}

func (c Config) RenderDefaultFilterChain() (*pbEnvoyListenerV3.FilterChain, error) {
	filters, err := c.RenderFilters()
	if err != nil {
		return nil, err
	}

	ret := &pbEnvoyListenerV3.FilterChain{
		Filters: filters,
	}

	if tls, err := c.DefaultTLS.RenderListenerTransportSocket(); err != nil {
		return nil, err
	} else {
		ret.TransportSocket = tls
	}

	return ret, nil
}

func (c Config) RenderSecondaryFilterChains() ([]*pbEnvoyListenerV3.FilterChain, error) {
	var r []*pbEnvoyListenerV3.FilterChain

	if chain, err := c.HttpToHttpsChain(); err != nil {
		return nil, err
	} else if chain != nil {
		r = append(r, chain)
	}

	if len(c.SNI) > 0 {
		filters, err := c.RenderFilters()
		if err != nil {
			return nil, err
		}

		chain, err := c.SNI.RenderFilterChain(filters)
		if err != nil {
			return nil, err
		}

		r = append(r, chain...)
	}

	return r, nil
}

func (c Config) HttpToHttpsChain() (*pbEnvoyListenerV3.FilterChain, error) {
	if c.DefaultTLS == nil {
		return nil, nil
	}

	httpFilterConfigType, err := anypb.New(&routerAPI.Router{})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create router filter configuration for HTTP to HTTPS redirect")
	}

	filterConfigType, err := anypb.New(&httpConnectionManagerAPI.HttpConnectionManager{
		StatPrefix: "ingress_http",
		CodecType:  httpConnectionManagerAPI.HttpConnectionManager_AUTO,
		RouteSpecifier: &httpConnectionManagerAPI.HttpConnectionManager_RouteConfig{
			RouteConfig: &pbEnvoyRouteV3.RouteConfiguration{
				Name:                           "local_http",
				MaxDirectResponseBodySizeBytes: wrapperspb.UInt32(utilConstants.MaxInventorySize),
				VirtualHosts: []*pbEnvoyRouteV3.VirtualHost{
					{
						Name:    "local_http",
						Domains: []string{"*"},
						Routes: []*pbEnvoyRouteV3.Route{
							{
								Match: &pbEnvoyRouteV3.RouteMatch{
									PathSpecifier: &pbEnvoyRouteV3.RouteMatch_Prefix{
										Prefix: "/",
									},
								},
								Action: &pbEnvoyRouteV3.Route_Redirect{
									Redirect: &pbEnvoyRouteV3.RedirectAction{
										SchemeRewriteSpecifier: &pbEnvoyRouteV3.RedirectAction_HttpsRedirect{
											HttpsRedirect: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		HttpFilters: []*httpConnectionManagerAPI.HttpFilter{
			{
				Name: "envoy.filters.http.router",
				ConfigType: &httpConnectionManagerAPI.HttpFilter_TypedConfig{
					TypedConfig: httpFilterConfigType,
				},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create HTTP connection manager configuration for HTTP to HTTPS redirect")
	}

	return &pbEnvoyListenerV3.FilterChain{
		FilterChainMatch: &pbEnvoyListenerV3.FilterChainMatch{
			TransportProtocol: "raw_buffer",
		},
		Filters: []*pbEnvoyListenerV3.Filter{
			{
				Name: "envoy.filters.network.http_connection_manager",
				ConfigType: &pbEnvoyListenerV3.Filter_TypedConfig{
					TypedConfig: filterConfigType,
				},
			},
		},
	}, nil
}

func (c Config) RenderListener() (*pbEnvoyListenerV3.Listener, error) {
	filterChains, err := c.RenderSecondaryFilterChains()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render secondary filter chains")
	}

	defaultFilterChain, err := c.RenderDefaultFilterChain()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render default filter")
	}

	var listenerFilters []*pbEnvoyListenerV3.ListenerFilter

	if c.DefaultTLS != nil {
		w, err := anypb.New(&tlsInspectorApi.TlsInspector{})
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to render TLS Inspector")
		}

		listenerFilters = append(listenerFilters, &pbEnvoyListenerV3.ListenerFilter{
			Name: "envoy.filters.listener.tls_inspector",
			ConfigType: &pbEnvoyListenerV3.ListenerFilter_TypedConfig{
				TypedConfig: w,
			},
		})
	}

	return &pbEnvoyListenerV3.Listener{
		Name: "default",
		Address: &pbEnvoyCoreV3.Address{
			Address: &pbEnvoyCoreV3.Address_SocketAddress{
				SocketAddress: &pbEnvoyCoreV3.SocketAddress{
					Address:       "0.0.0.0",
					PortSpecifier: &pbEnvoyCoreV3.SocketAddress_PortValue{PortValue: shared.ArangoPort},
				},
			},
		},
		FilterChains:       filterChains,
		ListenerFilters:    listenerFilters,
		DefaultFilterChain: defaultFilterChain,
	}, nil
}
