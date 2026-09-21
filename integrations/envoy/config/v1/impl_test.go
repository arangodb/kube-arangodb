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
	"os"
	"path/filepath"
	"testing"
	"time"

	pbEnvoyClusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	pbEnvoyListenerV3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	pbEnvoyConfigV1 "github.com/arangodb/kube-arangodb/integrations/envoy/config/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func newTestImpl(t *testing.T) *impl {
	// No local config files: the server starts with an empty snapshot until the operator pushes.
	h, err := New("", "", "")
	require.NoError(t, err)
	return h.(*impl)
}

// superuserContext returns a context carrying an authenticated superuser (server) identity: present, no user.
func superuserContext() context.Context {
	return authenticator.WithIdentity(context.Background(), &authenticator.Identity{})
}

func clusterAny(t *testing.T, name string) *anypb.Any {
	a, err := anypb.New(&pbEnvoyClusterV3.Cluster{
		Name:                 name,
		ConnectTimeout:       durationpb.New(time.Second),
		ClusterDiscoveryType: &pbEnvoyClusterV3.Cluster_Type{Type: pbEnvoyClusterV3.Cluster_STATIC},
	})
	require.NoError(t, err)
	return a
}

func Test_XDS_LocalSnapshot(t *testing.T) {
	dir := t.TempDir()
	cdsFile := filepath.Join(dir, "cds.yaml")
	ldsFile := filepath.Join(dir, "lds.yaml")
	versionFile := filepath.Join(dir, "version")

	cds, err := ugrpc.MarshalYAML(&discoveryservice.DiscoveryResponse{Resources: []*anypb.Any{clusterAny(t, "local_cluster")}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cdsFile, cds, 0644))

	listenerAny, err := anypb.New(&pbEnvoyListenerV3.Listener{Name: "local_listener"})
	require.NoError(t, err)
	lds, err := ugrpc.MarshalYAML(&discoveryservice.DiscoveryResponse{Resources: []*anypb.Any{listenerAny}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(ldsFile, lds, 0644))

	require.NoError(t, os.WriteFile(versionFile, []byte("checksum-123\n"), 0644))

	h, err := New(cdsFile, ldsFile, versionFile)
	require.NoError(t, err)
	i := h.(*impl)

	snap, err := i.cache.GetSnapshot(nodeKey)
	require.NoError(t, err)
	// The snapshot is seeded from the local files, versioned by the checksum (so the operator skips a
	// redundant push when unchanged).
	require.Equal(t, "checksum-123", snap.GetVersion(resourcev3.ClusterType))
	require.Contains(t, snap.GetResources(resourcev3.ClusterType), "local_cluster")
	require.Contains(t, snap.GetResources(resourcev3.ListenerType), "local_listener")
}

func Test_XDS_EmptySnapshot(t *testing.T) {
	i := newTestImpl(t)

	snap, err := i.cache.GetSnapshot(nodeKey)
	require.NoError(t, err)
	require.Equal(t, "empty", snap.GetVersion(resourcev3.ClusterType))
	require.Empty(t, snap.GetResources(resourcev3.ClusterType))
	require.Empty(t, snap.GetResources(resourcev3.ListenerType))
}

func Test_XDS_SetSnapshot(t *testing.T) {
	i := newTestImpl(t)

	require.NoError(t, i.SetSnapshot("v2", map[resourcev3.Type][]cachetypes.Resource{
		resourcev3.ClusterType: {
			&pbEnvoyClusterV3.Cluster{
				Name:                 "other_cluster",
				ConnectTimeout:       durationpb.New(time.Second),
				ClusterDiscoveryType: &pbEnvoyClusterV3.Cluster_Type{Type: pbEnvoyClusterV3.Cluster_STATIC},
			},
		},
	}))

	snap, err := i.cache.GetSnapshot(nodeKey)
	require.NoError(t, err)
	require.Equal(t, "v2", snap.GetVersion(resourcev3.ClusterType))
	require.Contains(t, snap.GetResources(resourcev3.ClusterType), "other_cluster")
	require.NotContains(t, snap.GetResources(resourcev3.ClusterType), "mock_cluster")
}

func Test_XDS_Push_RequiresSuperuser(t *testing.T) {
	i := newTestImpl(t)

	req := &pbEnvoyConfigV1.EnvoyConfigV1PushRequest{
		Version:  "v1",
		Clusters: []*anypb.Any{clusterAny(t, "pushed_cluster")},
	}

	t.Run("Unauthenticated is rejected", func(t *testing.T) {
		_, err := i.Push(context.Background(), req)
		tgrpc.AsGRPCError(t, err).Code(t, codes.Unauthenticated)
	})

	t.Run("Named user is rejected", func(t *testing.T) {
		ctx := authenticator.WithIdentity(context.Background(), &authenticator.Identity{User: util.NewType("root")})
		_, err := i.Push(ctx, req)
		tgrpc.AsGRPCError(t, err).Code(t, codes.PermissionDenied)
	})
}

func Test_XDS_Push_InstallsSnapshot(t *testing.T) {
	i := newTestImpl(t)
	ctx := superuserContext()

	resp, err := i.Push(ctx, &pbEnvoyConfigV1.EnvoyConfigV1PushRequest{
		Version:  "rev-42",
		Clusters: []*anypb.Any{clusterAny(t, "pushed_cluster")},
	})
	require.NoError(t, err)
	require.Equal(t, "rev-42", resp.GetVersion())

	// The pushed resources replace the mock snapshot.
	snap, err := i.cache.GetSnapshot(nodeKey)
	require.NoError(t, err)
	require.Equal(t, "rev-42", snap.GetVersion(resourcev3.ClusterType))
	require.Contains(t, snap.GetResources(resourcev3.ClusterType), "pushed_cluster")
	require.NotContains(t, snap.GetResources(resourcev3.ClusterType), "mock_cluster")

	// Status reports the served revision so the operator can push only on change.
	st, err := i.Status(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)
	require.Equal(t, "rev-42", st.GetVersion())
}
