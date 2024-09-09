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
	"path"

	bootstrapAPI "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	coreAPI "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryApi "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type DynamicConfig struct {
	Path, File string
}

func (d *DynamicConfig) AsConfigSource() *coreAPI.ConfigSource {
	if d == nil {
		return nil
	}

	return &coreAPI.ConfigSource{
		ConfigSourceSpecifier: &coreAPI.ConfigSource_PathConfigSource{
			PathConfigSource: &coreAPI.PathConfigSource{
				Path: path.Join(d.Path, d.File),
				WatchedDirectory: &coreAPI.WatchedDirectory{
					Path: d.Path,
				},
			},
		},
	}
}

func NodeDynamicConfig(cluster, id string, cds, lds *DynamicConfig) ([]byte, string, *bootstrapAPI.Bootstrap, error) {
	var b = bootstrapAPI.Bootstrap{
		Node: &coreAPI.Node{
			Id:      id,
			Cluster: cluster,
		},
	}

	if v := cds; v != nil {
		if b.DynamicResources == nil {
			b.DynamicResources = &bootstrapAPI.Bootstrap_DynamicResources{}
		}

		b.DynamicResources.CdsConfig = v.AsConfigSource()
	}

	if v := lds; v != nil {
		if b.DynamicResources == nil {
			b.DynamicResources = &bootstrapAPI.Bootstrap_DynamicResources{}
		}

		b.DynamicResources.LdsConfig = v.AsConfigSource()
	}

	return Marshal(&b)
}

func DynamicConfigResponse[T proto.Message](in ...T) (*discoveryApi.DiscoveryResponse, error) {
	resources := make([]*anypb.Any, len(in))
	for id := range in {
		if a, err := anypb.New(in[id]); err != nil {
			return nil, err
		} else {
			resources[id] = a
		}
	}
	return &discoveryApi.DiscoveryResponse{
		Resources: resources,
	}, nil
}
