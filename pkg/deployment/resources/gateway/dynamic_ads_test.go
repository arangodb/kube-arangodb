//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	"testing"

	pbEnvoyClusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/stretchr/testify/require"
)

func Test_NodeADSConfig(t *testing.T) {
	const adsCluster = "gateway_ads"

	data, sha, b, err := NodeADSConfig("arangodb", "gtw-1", adsCluster, ConfigDestinationTargetUnix{Path: "/var/run/sidecar/socket/api.sock"})
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.NotEmpty(t, sha)

	// Node identity.
	require.Equal(t, "gtw-1", b.GetNode().GetId())
	require.Equal(t, "arangodb", b.GetNode().GetCluster())

	// A single static cluster targets the ADS endpoint over HTTP/2.
	clusters := b.GetStaticResources().GetClusters()
	require.Len(t, clusters, 1)
	require.Equal(t, adsCluster, clusters[0].GetName())
	require.Equal(t, pbEnvoyClusterV3.Cluster_STATIC, clusters[0].GetType())
	require.Contains(t, clusters[0].GetTypedExtensionProtocolOptions(), "envoy.extensions.upstreams.http.v3.HttpProtocolOptions")
	require.Equal(t, "/var/run/sidecar/socket/api.sock", clusters[0].GetLoadAssignment().GetEndpoints()[0].GetLbEndpoints()[0].GetEndpoint().GetAddress().GetPipe().GetPath())

	// dynamic_resources subscribe via ADS to the static cluster.
	dr := b.GetDynamicResources()
	require.NotNil(t, dr.GetAdsConfig())
	require.Equal(t, adsCluster, dr.GetAdsConfig().GetGrpcServices()[0].GetEnvoyGrpc().GetClusterName())

	// CDS and LDS both resolve over the ADS stream.
	require.NotNil(t, dr.GetCdsConfig().GetAds())
	require.NotNil(t, dr.GetLdsConfig().GetAds())
}
