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
	routeAPI "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigDestinationsUpgrade []ConfigDestinationUpgrade

func (c ConfigDestinationsUpgrade) render() []*routeAPI.RouteAction_UpgradeConfig {
	if len(c) == 0 {
		return nil
	}

	var r = make([]*routeAPI.RouteAction_UpgradeConfig, len(c))

	for id := range c {
		r[id] = c[id].render()
	}

	return r
}

func (c ConfigDestinationsUpgrade) Validate() error {
	return shared.ValidateInterfaceList(c)
}

type ConfigDestinationUpgrade struct {
	Type string `json:"type"`

	Enabled *bool `json:"enabled,omitempty"`
}

func (c ConfigDestinationUpgrade) render() *routeAPI.RouteAction_UpgradeConfig {
	return &routeAPI.RouteAction_UpgradeConfig{
		UpgradeType: c.Type,
		Enabled:     wrapperspb.Bool(util.OptionalType(c.Enabled, true)),
	}
}

func (c ConfigDestinationUpgrade) Validate() error {
	switch c.Type {
	case "websocket":
		return nil
	default:
		return shared.PrefixResourceError("type", errors.Errorf("Unknown type: %s", c.Type))
	}
}
