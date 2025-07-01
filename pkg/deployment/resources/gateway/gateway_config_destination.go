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
	"time"

	pbEnvoyClusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	pbEnvoyEndpointV3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	pbEnvoyRouteV3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
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

	Match *ConfigMatch `json:"match,omitempty"`

	Path *string `json:"path,omitempty"`

	AuthExtension *ConfigAuthZExtension `json:"authExtension,omitempty"`

	HealthChecks ConfigDestinationHealthChecks `json:"healthChecks,omitempty"`

	UpgradeConfigs ConfigDestinationsUpgrade `json:"upgradeConfigs,omitempty"`

	TLS ConfigDestinationTLS `json:"tls,omitempty"`

	Timeout *meta.Duration `json:"timeout,omitempty"`

	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`

	Static ConfigDestinationStaticInterface `json:"static,omitempty"`
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
			shared.PrefixResourceError("pathType", shared.ValidateOptionalInterface(c.Match)),
			shared.PrefixResourceError("authExtension", c.AuthExtension.Validate()),
			shared.PrefixResourceError("static", shared.ValidateRequiredInterface(c.Static)),
		)
	default:
		return shared.WithErrors(
			shared.PrefixResourceError("targets", c.Targets.Validate()),
			shared.PrefixResourceError("type", c.Type.Validate()),
			shared.PrefixResourceError("protocol", c.Protocol.Validate()),
			shared.PrefixResourceError("tls", c.TLS.Validate()),
			shared.PrefixResourceError("healthChecks", c.HealthChecks.Validate()),
			shared.PrefixResourceError("path", shared.ValidateAPIPath(c.GetPath())),
			shared.PrefixResourceError("pathType", shared.ValidateOptionalInterface(c.Match)),
			shared.PrefixResourceError("authExtension", c.AuthExtension.Validate()),
			shared.PrefixResourceError("upgradeConfigs", c.UpgradeConfigs.Validate()),
			shared.PrefixResourceErrorFunc("timeout", func() error {
				if t := c.GetTimeout(); t < constants.MinEnvoyUpstreamTimeout {
					return errors.Errorf("Timeout lower than %s not allowed", constants.MinEnvoyUpstreamTimeout.String())
				} else if t > constants.MaxEnvoyUpstreamTimeout {
					return errors.Errorf("Timeout greater than %s not allowed", constants.MaxEnvoyUpstreamTimeout.String())
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

func (c *ConfigDestination) RenderRoute(name, prefix string) (*pbEnvoyRouteV3.Route, error) {
	if c == nil {
		return nil, errors.Errorf("Route cannot be nil")
	}
	var headers []*pbEnvoyCoreV3.HeaderValueOption

	for k, v := range c.ResponseHeaders {
		headers = append(headers, &pbEnvoyCoreV3.HeaderValueOption{
			Header: &pbEnvoyCoreV3.HeaderValue{
				Key:   k,
				Value: v,
			},
			AppendAction:   pbEnvoyCoreV3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			KeepEmptyValue: false,
		})
	}

	var tcg []TypedFilterConfigGen

	if c.AuthExtension != nil {
		tcg = append(tcg, c.AuthExtension)
	}
	tc, err := NewTypedFilterConfig(tcg...)
	if err != nil {
		return nil, err
	}

	r := &pbEnvoyRouteV3.Route{
		Match:                c.Match.Match(prefix),
		ResponseHeadersToAdd: headers,

		TypedPerFilterConfig: tc,
	}

	if err := c.appendRouteAction(r, name); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *ConfigDestination) appendRouteAction(route *pbEnvoyRouteV3.Route, name string) error {
	if c.Type.Get() == ConfigDestinationTypeStatic {
		if c.Static == nil {
			return errors.Errorf("Static response is not defined!")
		}
		data, code, err := c.Static.StaticResponse()
		if err != nil {
			return err
		}

		// Return static response
		route.Action = &pbEnvoyRouteV3.Route_DirectResponse{
			DirectResponse: &pbEnvoyRouteV3.DirectResponseAction{
				Status: code,
				Body: &pbEnvoyCoreV3.DataSource{
					Specifier: &pbEnvoyCoreV3.DataSource_InlineBytes{
						InlineBytes: data,
					},
				},
			},
		}
		return nil
	}

	route.Action = &pbEnvoyRouteV3.Route_Route{
		Route: &pbEnvoyRouteV3.RouteAction{
			ClusterSpecifier: &pbEnvoyRouteV3.RouteAction_Cluster{
				Cluster: name,
			},
			UpgradeConfigs: c.getUpgradeConfigs().render(),
			PrefixRewrite:  c.GetPath(),
			Timeout:        durationpb.New(c.GetTimeout()),
			IdleTimeout:    durationpb.New(c.GetTimeout()),
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

func (c *ConfigDestination) RenderCluster(name string) (*pbEnvoyClusterV3.Cluster, error) {
	if c.Type.Get() == ConfigDestinationTypeStatic {
		return nil, nil
	}

	hpo, err := anypb.New(c.Protocol.Options())
	if err != nil {
		return nil, err
	}

	cluster := &pbEnvoyClusterV3.Cluster{
		Name:           name,
		ConnectTimeout: durationpb.New(time.Second),
		LbPolicy:       pbEnvoyClusterV3.Cluster_ROUND_ROBIN,
		LoadAssignment: &pbEnvoyEndpointV3.ClusterLoadAssignment{
			ClusterName: name,
			Endpoints: []*pbEnvoyEndpointV3.LocalityLbEndpoints{
				{
					LbEndpoints: c.Targets.RenderEndpoints(),
				},
			},
		},
		HealthChecks: c.HealthChecks.Render(),
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
