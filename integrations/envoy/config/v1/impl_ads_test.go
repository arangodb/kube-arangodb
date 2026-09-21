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

package v1

import (
	"context"
	"net"
	"testing"
	"time"

	pbEnvoyClusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Test_XDS_FastReload proves that a config change propagates to a connected Envoy (ADS client) near
// instantly over the stream - not on a polling/sync interval - which is the point of push mode. It starts
// the real ADS server, connects a client, then measures the time from SetSnapshot to the client receiving
// the update.
func Test_XDS_FastReload(t *testing.T) {
	h, err := New("", "", "")
	require.NoError(t, err)
	i := h.(*impl)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	i.Register(srv)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	stream, err := discoveryservice.NewAggregatedDiscoveryServiceClient(conn).StreamAggregatedResources(ctx)
	require.NoError(t, err)

	node := &pbEnvoyCoreV3.Node{Id: "gtw-1"}

	// Subscribe to CDS and receive the initial (empty) snapshot.
	require.NoError(t, stream.Send(&discoveryservice.DiscoveryRequest{Node: node, TypeUrl: resourcev3.ClusterType}))

	resp1, err := stream.Recv()
	require.NoError(t, err)
	require.Equal(t, "empty", resp1.GetVersionInfo())
	require.Empty(t, resp1.GetResources())

	// ACK the initial snapshot so the server will push the next version.
	require.NoError(t, stream.Send(&discoveryservice.DiscoveryRequest{
		Node:          node,
		TypeUrl:       resourcev3.ClusterType,
		VersionInfo:   resp1.GetVersionInfo(),
		ResponseNonce: resp1.GetNonce(),
	}))

	// Push a new config and measure how quickly the client receives it.
	newCluster := &pbEnvoyClusterV3.Cluster{
		Name:                 "fast_cluster",
		ConnectTimeout:       durationpb.New(time.Second),
		ClusterDiscoveryType: &pbEnvoyClusterV3.Cluster_Type{Type: pbEnvoyClusterV3.Cluster_STATIC},
	}

	start := time.Now()
	require.NoError(t, i.SetSnapshot("fast-v2", map[resourcev3.Type][]cachetypes.Resource{
		resourcev3.ClusterType: {newCluster},
	}))

	resp2, err := stream.Recv()
	elapsed := time.Since(start)
	require.NoError(t, err)

	require.Equal(t, "fast-v2", resp2.GetVersionInfo())

	// The push must reach the client near-instantly over the ADS stream, not after a sync/poll interval.
	require.Less(t, elapsed, 2*time.Second, "ADS push propagation took too long: %s", elapsed)

	t.Logf("ADS config push propagated to the client in %s", elapsed)
}
