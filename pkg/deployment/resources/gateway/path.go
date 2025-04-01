//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigMatch int

const (
	ConfigMatchPrefix ConfigMatch = iota
	ConfigMatchPath
)

func (c *ConfigMatch) Get() ConfigMatch {
	if c == nil {
		return ConfigMatchPrefix
	}

	switch v := *c; v {
	case ConfigMatchPrefix, ConfigMatchPath:
		return v
	default:
		return ConfigMatchPrefix
	}
}

func (c *ConfigMatch) Validate() error {
	switch c.Get() {
	case ConfigMatchPrefix, ConfigMatchPath:
		return nil
	default:
		return errors.Errorf("Invalid path type")
	}
}

func (c *ConfigMatch) Match(path string) *routeAPI.RouteMatch {
	switch c.Get() {
	case ConfigMatchPath:
		return &routeAPI.RouteMatch{
			PathSpecifier: &routeAPI.RouteMatch_Path{
				Path: path,
			},
		}
	default:
		return &routeAPI.RouteMatch{
			PathSpecifier: &routeAPI.RouteMatch_Prefix{
				Prefix: path,
			},
		}
	}
}
