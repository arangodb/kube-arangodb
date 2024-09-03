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

	clusterAPI "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointAPI "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	routeAPI "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
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

	Path *string `json:"path,omitempty"`

	AuthExtension *ConfigAuthZExtension `json:"authExtension,omitempty"`
}

func (c *ConfigDestination) Validate() error {
	if c == nil {
		c = &ConfigDestination{}
	}
	return shared.WithErrors(
		shared.PrefixResourceError("targets", c.Targets.Validate()),
		shared.PrefixResourceError("type", c.Type.Validate()),
		shared.PrefixResourceError("path", shared.ValidateAPIPath(c.GetPath())),
		shared.PrefixResourceError("authExtension", c.AuthExtension.Validate()),
	)
}

func (c *ConfigDestination) GetPath() string {
	if c == nil || c.Path == nil {
		return "/"
	}

	return *c.Path
}

func (c *ConfigDestination) RenderRoute(name, prefix string) (*routeAPI.Route, error) {
	var tcg []TypedFilterConfigGen

	if c != nil && c.AuthExtension != nil {
		tcg = append(tcg, c.AuthExtension)
	}
	tc, err := NewTypedFilterConfig(tcg...)
	if err != nil {
		return nil, err
	}

	return &routeAPI.Route{
		Match: &routeAPI.RouteMatch{
			PathSpecifier: &routeAPI.RouteMatch_Prefix{
				Prefix: prefix,
			},
		},
		Action: &routeAPI.Route_Route{
			Route: &routeAPI.RouteAction{
				ClusterSpecifier: &routeAPI.RouteAction_Cluster{
					Cluster: name,
				},
				PrefixRewrite: c.GetPath(),
			},
		},
		TypedPerFilterConfig: tc,
	}, nil
}

func (c *ConfigDestination) RenderCluster(name string) (*clusterAPI.Cluster, error) {
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
	}

	if t, err := c.Type.RenderUpstreamTransportSocket(); err != nil {
		return nil, err
	} else {
		cluster.TransportSocket = t
	}

	return cluster, nil
}
