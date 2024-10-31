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
	listenerAPI "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigSNIList []ConfigSNI

func (c ConfigSNIList) RenderFilterChain(filters []*listenerAPI.Filter) ([]*listenerAPI.FilterChain, error) {
	var r = make([]*listenerAPI.FilterChain, len(c))
	for id := range c {
		if f, err := c[id].RenderFilterChain(filters); err != nil {
			return nil, err
		} else {
			r[id] = f
		}
	}
	return r, nil
}

func (c ConfigSNIList) Validate() error {
	return shared.ValidateList(c, func(sni ConfigSNI) error {
		return sni.Validate()
	})
}

type ConfigSNI struct {
	ConfigTLS `json:",inline"`

	ServerNames []string `json:"serverNames,omitempty"`
}

func (c ConfigSNI) RenderFilterChain(filters []*listenerAPI.Filter) (*listenerAPI.FilterChain, error) {
	transport, err := c.RenderListenerTransportSocket()
	if err != nil {
		return nil, err
	}

	return &listenerAPI.FilterChain{
		TransportSocket: transport,
		FilterChainMatch: &listenerAPI.FilterChainMatch{
			ServerNames: util.CopyList(c.ServerNames),
		},
		Filters: filters,
	}, nil
}

func (c ConfigSNI) Validate() error {
	return shared.WithErrors(
		shared.ValidateList(c.ServerNames, sharedApi.IsValidDomain, func(in []string) error {
			if len(in) == 0 {
				return errors.Errorf("AtLeast one element required on list")
			}
			return nil
		}),
	)
}
