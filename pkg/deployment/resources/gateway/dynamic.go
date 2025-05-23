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
	"path"

	pbEnvoyBootstrapV3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryApi "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/arangodb/kube-arangodb/pkg/util"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

type DynamicConfig struct {
	Path, File string
}

func (d *DynamicConfig) AsConfigSource() *pbEnvoyCoreV3.ConfigSource {
	if d == nil {
		return nil
	}

	return &pbEnvoyCoreV3.ConfigSource{
		ConfigSourceSpecifier: &pbEnvoyCoreV3.ConfigSource_PathConfigSource{
			PathConfigSource: &pbEnvoyCoreV3.PathConfigSource{
				Path: path.Join(d.Path, d.File),
				WatchedDirectory: &pbEnvoyCoreV3.WatchedDirectory{
					Path: d.Path,
				},
			},
		},
	}
}

func NodeDynamicConfig(cluster, id string, cds, lds *DynamicConfig) ([]byte, string, *pbEnvoyBootstrapV3.Bootstrap, error) {
	var b = pbEnvoyBootstrapV3.Bootstrap{
		Node: &pbEnvoyCoreV3.Node{
			Id:      id,
			Cluster: cluster,
		},
	}

	if v := cds; v != nil {
		if b.DynamicResources == nil {
			b.DynamicResources = &pbEnvoyBootstrapV3.Bootstrap_DynamicResources{}
		}

		b.DynamicResources.CdsConfig = v.AsConfigSource()
	}

	if v := lds; v != nil {
		if b.DynamicResources == nil {
			b.DynamicResources = &pbEnvoyBootstrapV3.Bootstrap_DynamicResources{}
		}

		b.DynamicResources.LdsConfig = v.AsConfigSource()
	}

	data, err := ugrpc.MarshalYAML(&b)
	if err != nil {
		return nil, "", nil, err
	}

	return data, util.SHA256(data), &b, nil
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
