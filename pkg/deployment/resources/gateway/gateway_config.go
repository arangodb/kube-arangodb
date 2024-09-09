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
	"fmt"
	"sort"
	"time"

	bootstrapAPI "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterAPI "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	coreAPI "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointAPI "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerAPI "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routeAPI "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	httpFilterAuthzApi "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	routerAPI "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	tlsInspectorApi "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	httpConnectionManagerAPI "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	upstreamHttpApi "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	discoveryApi "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Config struct {
	DefaultDestination ConfigDestination `json:"defaultDestination,omitempty"`

	Destinations ConfigDestinations `json:"destinations,omitempty"`

	DefaultTLS *ConfigTLS `json:"defaultTLS,omitempty"`

	IntegrationSidecar *ConfigDestinationTarget `json:"integrationSidecar,omitempty"`

	SNI ConfigSNIList `json:"sni,omitempty"`
}

func (c Config) Validate() error {
	return errors.Errors(
		shared.PrefixResourceErrors("defaultDestination", c.DefaultDestination.Validate()),
		shared.PrefixResourceErrors("integrationSidecar", c.IntegrationSidecar.Validate()),
		shared.PrefixResourceErrors("destinations", c.Destinations.Validate()),
		shared.PrefixResourceErrors("sni", c.SNI.Validate()),
	)
}

func (c Config) RenderYAML() ([]byte, string, *bootstrapAPI.Bootstrap, error) {
	cfg, err := c.Render()
	if err != nil {
		return nil, "", nil, err
	}

	return Marshal(cfg)
}

func (c Config) RenderCDSYAML() ([]byte, string, *discoveryApi.DiscoveryResponse, error) {
	cfg, err := c.RenderCDS()
	if err != nil {
		return nil, "", nil, err
	}

	return Marshal(cfg)
}

func (c Config) RenderLDSYAML() ([]byte, string, *discoveryApi.DiscoveryResponse, error) {
	cfg, err := c.RenderLDS()
	if err != nil {
		return nil, "", nil, err
	}

	return Marshal(cfg)
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

func (c Config) Render() (*bootstrapAPI.Bootstrap, error) {
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

func (c Config) RenderClusters() ([]*clusterAPI.Cluster, error) {
	def, err := c.DefaultDestination.RenderCluster("default")
	if err != nil {
		return nil, err
	}
	clusters := []*clusterAPI.Cluster{
		def,
	}

	if i := c.IntegrationSidecar; i != nil {
		hpo, err := anypb.New(&upstreamHttpApi.HttpProtocolOptions{
			UpstreamProtocolOptions: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &upstreamHttpApi.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
						Http2ProtocolOptions: &coreAPI.Http2ProtocolOptions{},
					},
				},
			},
		})
		if err != nil {
			return nil, err
		}
		cluster := &clusterAPI.Cluster{
			Name:           "integration_sidecar",
			ConnectTimeout: durationpb.New(time.Second),
			LbPolicy:       clusterAPI.Cluster_ROUND_ROBIN,
			LoadAssignment: &endpointAPI.ClusterLoadAssignment{
				ClusterName: "integration_sidecar",
				Endpoints: []*endpointAPI.LocalityLbEndpoints{
					{
						LbEndpoints: []*endpointAPI.LbEndpoint{
							i.RenderEndpoint(),
						},
					},
				},
			},
			TypedExtensionProtocolOptions: map[string]*any.Any{
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

		clusters = append(clusters, c)
	}

	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Name < clusters[j].Name
	})

	return clusters, nil
}

func (c Config) RenderRoutes() ([]*routeAPI.Route, error) {
	def, err := c.DefaultDestination.RenderRoute("default", "/")
	if err != nil {
		return nil, err
	}
	routes := []*routeAPI.Route{
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

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].GetMatch().GetPrefix() > routes[j].GetMatch().GetPrefix()
	})

	return routes, nil
}

func (c Config) RenderIntegrationSidecarFilter() (*httpConnectionManagerAPI.HttpFilter, error) {
	e, err := anypb.New(&httpFilterAuthzApi.ExtAuthz{
		Services: &httpFilterAuthzApi.ExtAuthz_GrpcService{
			GrpcService: &coreAPI.GrpcService{
				TargetSpecifier: &coreAPI.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &coreAPI.GrpcService_EnvoyGrpc{
						ClusterName: "integration_sidecar",
					},
				},
				Timeout: durationpb.New(500 * time.Millisecond),
			},
		},
		IncludePeerCertificate: true,
	})
	if err != nil {
		return nil, err
	}

	return &httpConnectionManagerAPI.HttpFilter{
		Name: IntegrationSidecarFilterName,
		ConfigType: &httpConnectionManagerAPI.HttpFilter_TypedConfig{
			TypedConfig: e,
		},
		IsOptional: false,
	}, nil
}

func (c Config) RenderFilters() ([]*listenerAPI.Filter, error) {
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
		StatPrefix: "ingress_http",
		CodecType:  httpConnectionManagerAPI.HttpConnectionManager_AUTO,
		RouteSpecifier: &httpConnectionManagerAPI.HttpConnectionManager_RouteConfig{
			RouteConfig: &routeAPI.RouteConfiguration{
				Name: "default",
				VirtualHosts: []*routeAPI.VirtualHost{
					{
						Name:    "default",
						Domains: []string{"*"},
						Routes:  routes,
					},
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

	return []*listenerAPI.Filter{
		{
			Name: "envoy.filters.network.httpConnectionManagerAPI",
			ConfigType: &listenerAPI.Filter_TypedConfig{
				TypedConfig: filterConfigType,
			},
		},
	}, nil
}

func (c Config) RenderDefaultFilterChain() (*listenerAPI.FilterChain, error) {
	filters, err := c.RenderFilters()
	if err != nil {
		return nil, err
	}

	ret := &listenerAPI.FilterChain{
		Filters: filters,
	}

	if tls, err := c.DefaultTLS.RenderListenerTransportSocket(); err != nil {
		return nil, err
	} else {
		ret.TransportSocket = tls
	}

	return ret, nil
}

func (c Config) RenderSecondaryFilterChains() ([]*listenerAPI.FilterChain, error) {
	if len(c.SNI) == 0 {
		return nil, nil
	}

	filters, err := c.RenderFilters()
	if err != nil {
		return nil, err
	}

	return c.SNI.RenderFilterChain(filters)
}

func (c Config) RenderListener() (*listenerAPI.Listener, error) {
	filterChains, err := c.RenderSecondaryFilterChains()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render secondary filter chains")
	}

	defaultFilterChain, err := c.RenderDefaultFilterChain()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to render default filter")
	}

	var listenerFilters []*listenerAPI.ListenerFilter

	if c.DefaultTLS != nil {
		w, err := anypb.New(&tlsInspectorApi.TlsInspector{})
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to render TLS Inspector")
		}

		listenerFilters = append(listenerFilters, &listenerAPI.ListenerFilter{
			Name: "envoy.filters.listener.tls_inspector",
			ConfigType: &listenerAPI.ListenerFilter_TypedConfig{
				TypedConfig: w,
			},
		})
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
		FilterChains:       filterChains,
		ListenerFilters:    listenerFilters,
		DefaultFilterChain: defaultFilterChain,
	}, nil
}
