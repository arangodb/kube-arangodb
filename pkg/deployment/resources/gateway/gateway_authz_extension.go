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
	httpFilterAuthzApi "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	"google.golang.org/protobuf/types/known/anypb"

	pbImplEnvoyAuthV3 "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type ConfigAuthZExtension struct {
	AuthZExtension map[string]string `json:"authZExtension,omitempty"`
}

func (c *ConfigAuthZExtension) RenderTypedFilterConfig() (util.KV[string, *anypb.Any], error) {
	if c == nil {
		return util.KV[string, *anypb.Any]{}, nil
	}

	var data = map[string]string{}

	for k, v := range c.AuthZExtension {
		data[k] = v
	}

	data[pbImplEnvoyAuthV3.AuthConfigTypeKey] = pbImplEnvoyAuthV3.AuthConfigTypeValue

	q, err := anypb.New(&httpFilterAuthzApi.ExtAuthzPerRoute{
		Override: &httpFilterAuthzApi.ExtAuthzPerRoute_CheckSettings{
			CheckSettings: &httpFilterAuthzApi.CheckSettings{
				ContextExtensions: data,
			},
		},
	})
	if err != nil {
		return util.KV[string, *anypb.Any]{}, err
	}

	return util.KV[string, *anypb.Any]{
		K: IntegrationSidecarFilterName,
		V: q,
	}, nil
}

func (c *ConfigAuthZExtension) Validate() error {
	return nil
}
