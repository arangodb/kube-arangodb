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
	"encoding/json"
	"time"

	clusterAPI "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	coreAPI "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointAPI "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	routeAPI "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigDestinations map[string]ConfigDestination

func (c ConfigDestinations) Validate() error {
	if len(c) == 0 {
		return nil
	}
	return shared.WithErrors(
		shared.ValidateMap(c, func(k string, destination ConfigDestination) error {
			var errs []error
			if k == "/" {
				errs = append(errs, errors.Errorf("Route for `/` is reserved"))
			}
			if err := shared.ValidateAPIPath(k); err != nil {
				errs = append(errs, err)
			}
			if err := destination.Validate(); err != nil {
				errs = append(errs, err)
			}
			return shared.WithErrors(errs...)
		}),
	)
}

type ConfigDestination struct {
	Targets ConfigDestinationTargets `json:"targets,omitempty"`

	Type *ConfigDestinationType `json:"type,omitempty"`

	Protocol *ConfigDestinationProtocol `json:"protocol,omitempty"`

	Path *string `json:"path,omitempty"`

	AuthExtension *ConfigAuthZExtension `json:"authExtension,omitempty"`

	UpgradeConfigs ConfigDestinationsUpgrade `json:"upgradeConfigs,omitempty"`

	TLS ConfigDestinationTLS `json:"tls,omitempty"`

	Timeout *meta.Duration `json:"timeout,omitempty"`

	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`

	Static *ConfigDestinationStatic `json:"static,omitempty"`
}

func (c *ConfigDestination) Validate() error {
	if c == nil {
		c = &ConfigDestination{}
	}

	switch c.Type.Get() {
	case ConfigDestinationTypeStatic:
		return shared.WithErrors(
			shared.PrefixResourceError("type", c.Type.Validate()),
			shared.PrefixResourceError("path", shared.ValidateAPIPath(c.GetPath())),
			shared.PrefixResourceError("authExtension", c.AuthExtension.Validate()),
			shared.PrefixResourceError("static", shared.ValidateRequiredInterface(c.Static)),
		)
	default:
		return shared.WithErrors(
			shared.PrefixResourceError("targets", c.Targets.Validate()),
			shared.PrefixResourceError("type", c.Type.Validate()),
			shared.PrefixResourceError("protocol", c.Protocol.Validate()),
			shared.PrefixResourceError("tls", c.TLS.Validate()),
			shared.PrefixResourceError("path", shared.ValidateAPIPath(c.GetPath())),
			shared.PrefixResourceError("authExtension", c.AuthExtension.Validate()),
			shared.PrefixResourceError("upgradeConfigs", c.UpgradeConfigs.Validate()),
			shared.PrefixResourceErrorFunc("timeout", func() error {
				if t := c.GetTimeout(); t < 15*time.Second {
					return errors.Errorf("Timeout lower than 15 seconds not allowed")
				} else if t > 15*time.Minute {
					return errors.Errorf("Timeout greater than 15 seconds not allowed")
				}
				return nil
			}),
		)
	}
}

func (c *ConfigDestination) GetTimeout() time.Duration {
	if c == nil || c.Timeout == nil {
		return constants.DefaultEnvoyUpstreamTimeout
	}

	return c.Timeout.Duration
}

func (c *ConfigDestination) GetPath() string {
	if c == nil || c.Path == nil {
		return "/"
	}

	return *c.Path
}

func (c *ConfigDestination) RenderRoute(name, prefix string) (*routeAPI.Route, error) {
	var headers []*coreAPI.HeaderValueOption

	for k, v := range c.ResponseHeaders {
		headers = append(headers, &coreAPI.HeaderValueOption{
			Header: &coreAPI.HeaderValue{
				Key:   k,
				Value: v,
			},
			AppendAction:   coreAPI.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			KeepEmptyValue: false,
		})
	}

	var tcg []TypedFilterConfigGen

	if c != nil && c.AuthExtension != nil {
		tcg = append(tcg, c.AuthExtension)
	}
	tc, err := NewTypedFilterConfig(tcg...)
	if err != nil {
		return nil, err
	}

	r := &routeAPI.Route{
		Match: &routeAPI.RouteMatch{
			PathSpecifier: &routeAPI.RouteMatch_Prefix{
				Prefix: prefix,
			},
		},
		ResponseHeadersToAdd: headers,

		TypedPerFilterConfig: tc,
	}

	if err := c.appendRouteAction(r, name); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *ConfigDestination) appendRouteAction(route *routeAPI.Route, name string) error {
	if c.Type.Get() == ConfigDestinationTypeStatic {
		obj := c.Static.GetResponse()

		if obj == nil {
			obj = struct{}{}
		}

		data, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		// Return static response
		route.Action = &routeAPI.Route_DirectResponse{
			DirectResponse: &routeAPI.DirectResponseAction{
				Status: c.Static.GetCode(),
				Body: &coreAPI.DataSource{
					Specifier: &coreAPI.DataSource_InlineBytes{
						InlineBytes: data,
					},
				},
			},
		}
		return nil
	}

	route.Action = &routeAPI.Route_Route{
		Route: &routeAPI.RouteAction{
			ClusterSpecifier: &routeAPI.RouteAction_Cluster{
				Cluster: name,
			},
			UpgradeConfigs: c.getUpgradeConfigs().render(),
			PrefixRewrite:  c.GetPath(),
			Timeout:        durationpb.New(c.GetTimeout()),
		},
	}
	return nil
}

func (c *ConfigDestination) getUpgradeConfigs() ConfigDestinationsUpgrade {
	if c == nil {
		return nil
	}

	return c.UpgradeConfigs
}

func (c *ConfigDestination) RenderCluster(name string) (*clusterAPI.Cluster, error) {
	if c.Type.Get() == ConfigDestinationTypeStatic {
		return nil, nil
	}

	hpo, err := anypb.New(c.Protocol.Options())
	if err != nil {
		return nil, err
	}

	cluster := &clusterAPI.Cluster{
		Name:           name,
		ConnectTimeout: durationpb.New(time.Second),
		LbPolicy:       clusterAPI.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpointAPI.ClusterLoadAssignment{
			ClusterName: name,
			Endpoints: []*endpointAPI.LocalityLbEndpoints{
				{
					LbEndpoints: c.Targets.RenderEndpoints(),
				},
			},
		},
		TypedExtensionProtocolOptions: map[string]*anypb.Any{
			"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": hpo,
		},
	}

	if t, err := c.Type.RenderUpstreamTransportSocket(c.Protocol, c.TLS); err != nil {
		return nil, err
	} else {
		cluster.TransportSocket = t
	}

	return cluster, nil
}
