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
	goStrings "strings"
	"sync"

	pbEnvoyCoreV3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	pbEnvoyConfigV1 "github.com/arangodb/kube-arangodb/integrations/envoy/config/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

// nodeKey is the single snapshot key every gateway node is served from. The dynamic config is identical
// across gateway members (each Envoy renders the same CDS/LDS), so a constant key lets any connecting node
// receive the current snapshot.
const nodeKey = "gateway"

// constHash maps every Envoy node to the single snapshot key.
type constHash struct{}

func (constHash) ID(_ *pbEnvoyCoreV3.Node) string { return nodeKey }

// New builds the EnvoyConfigV1 xDS integration: an ADS server backed by a SnapshotCache. Envoy connects to
// it (as a cluster over the integration sidecar) and receives the config via ADS.
//
// The initial snapshot is loaded from the gateway CDS/LDS ConfigMap files mounted into the sidecar
// (cdsFile/ldsFile hold DiscoveryResponse YAML, versionFile the config checksum). This makes a restarted
// gateway serve the last-known-good local config immediately - without waiting for the operator to push -
// and, because the snapshot version equals the ConfigMap checksum, the operator skips a redundant push when
// the config is unchanged. If the local config cannot be read (files missing or unparsable), the server
// starts with an empty snapshot and relies on the operator control channel to push the config via Push.
func New(cdsFile, ldsFile, versionFile string) (svc.Handler, error) {
	cache := cachev3.NewSnapshotCache(true, constHash{}, gcpLogger{})

	i := &impl{
		ctx:   context.Background(),
		cache: cache,
	}

	i.server = serverv3.NewServer(i.ctx, cache, nil)

	version, resources, err := loadLocalSnapshot(cdsFile, ldsFile, versionFile)
	if err != nil {
		logger.Err(err).Warn("Unable to load local gateway config for the initial snapshot; starting empty and waiting for an operator push")
		version, resources = "empty", map[resourcev3.Type][]cachetypes.Resource{
			resourcev3.ClusterType:  {},
			resourcev3.ListenerType: {},
		}
	}

	if err := i.SetSnapshot(version, resources); err != nil {
		return nil, err
	}

	return i, nil
}

var _ svc.Handler = &impl{}
var _ pbEnvoyConfigV1.EnvoyConfigV1Server = &impl{}

type impl struct {
	pbEnvoyConfigV1.UnsafeEnvoyConfigV1Server

	ctx    context.Context
	cache  cachev3.SnapshotCache
	server serverv3.Server

	lock    sync.Mutex
	version string
}

func (i *impl) Name() string {
	return Name
}

func (i *impl) Health(ctx context.Context) svc.HealthState {
	return svc.Healthy
}

func (i *impl) Gateway(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return nil
}

func (i *impl) Register(registrar *grpc.Server) {
	// ADS server consumed by Envoy.
	discoveryservice.RegisterAggregatedDiscoveryServiceServer(registrar, i.server)
	// Operator control channel that feeds the snapshot.
	pbEnvoyConfigV1.RegisterEnvoyConfigV1Server(registrar, i)
}

// Push installs the operator-provided CDS/LDS resources as the current ADS snapshot. It requires a
// superuser (server) identity: the config controls how Envoy routes traffic, so only the operator (which
// authenticates with a superuser JWT) may set it.
func (i *impl) Push(ctx context.Context, request *pbEnvoyConfigV1.EnvoyConfigV1PushRequest) (*pbEnvoyConfigV1.EnvoyConfigV1PushResponse, error) {
	if err := requireSuperuser(ctx); err != nil {
		return nil, err
	}

	clusters, err := decodeResources(request.GetClusters())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to decode clusters: %s", err.Error())
	}

	listeners, err := decodeResources(request.GetListeners())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to decode listeners: %s", err.Error())
	}

	if err := i.SetSnapshot(request.GetVersion(), map[resourcev3.Type][]cachetypes.Resource{
		resourcev3.ClusterType:  clusters,
		resourcev3.ListenerType: listeners,
	}); err != nil {
		logger.Err(err).Str("version", request.GetVersion()).Warn("Unable to install pushed snapshot")
		return nil, status.Errorf(codes.InvalidArgument, "Unable to install snapshot: %s", err.Error())
	}

	logger.Str("version", request.GetVersion()).Int("clusters", len(clusters)).Int("listeners", len(listeners)).Info("Config snapshot pushed")

	return &pbEnvoyConfigV1.EnvoyConfigV1PushResponse{
		Version: request.GetVersion(),
	}, nil
}

// Status returns the version of the snapshot currently served, so the operator can push only on change.
func (i *impl) Status(ctx context.Context, _ *pbSharedV1.Empty) (*pbEnvoyConfigV1.EnvoyConfigV1StatusResponse, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	return &pbEnvoyConfigV1.EnvoyConfigV1StatusResponse{
		Version: i.version,
	}, nil
}

// SetSnapshot installs a new xDS snapshot at the given version. Envoy connected via ADS receives the
// updated resources without a restart.
func (i *impl) SetSnapshot(version string, resources map[resourcev3.Type][]cachetypes.Resource) error {
	snap, err := cachev3.NewSnapshot(version, resources)
	if err != nil {
		return err
	}

	if err := snap.Consistent(); err != nil {
		return err
	}

	if err := i.cache.SetSnapshot(i.ctx, nodeKey, snap); err != nil {
		return err
	}

	i.lock.Lock()
	i.version = version
	i.lock.Unlock()

	return nil
}

// requireSuperuser rejects the call unless the authenticated identity is a superuser (server) token: an
// authenticated identity with no username. A named user or an unauthenticated caller is denied.
func requireSuperuser(ctx context.Context) error {
	id := authenticator.GetIdentity(ctx)
	if id == nil {
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	}
	if id.User != nil {
		return status.Error(codes.PermissionDenied, "Superuser (server) identity required")
	}
	return nil
}

// decodeResources unmarshals the xDS resources carried as Any into their concrete proto messages, which
// the SnapshotCache requires to derive resource names.
func decodeResources(in []*anypb.Any) ([]cachetypes.Resource, error) {
	res := make([]cachetypes.Resource, 0, len(in))
	for _, a := range in {
		m, err := a.UnmarshalNew()
		if err != nil {
			return nil, err
		}

		r, ok := m.(cachetypes.Resource)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "Type %s is not a valid xDS resource", a.GetTypeUrl())
		}

		res = append(res, r)
	}
	return res, nil
}

// loadLocalSnapshot reads the gateway CDS and LDS DiscoveryResponse files (as mounted from the gateway
// ConfigMaps) into an xDS snapshot, versioned by the config checksum so it matches what the operator would
// push. It errors if either resource file cannot be read or parsed.
func loadLocalSnapshot(cdsFile, ldsFile, versionFile string) (string, map[resourcev3.Type][]cachetypes.Resource, error) {
	clusters, err := loadResourceFile(cdsFile)
	if err != nil {
		return "", nil, err
	}

	listeners, err := loadResourceFile(ldsFile)
	if err != nil {
		return "", nil, err
	}

	version := "local"
	if versionFile != "" {
		if data, err := os.ReadFile(versionFile); err == nil {
			if v := goStrings.TrimSpace(string(data)); v != "" {
				version = v
			}
		}
	}

	return version, map[resourcev3.Type][]cachetypes.Resource{
		resourcev3.ClusterType:  clusters,
		resourcev3.ListenerType: listeners,
	}, nil
}

// loadResourceFile reads a gateway CDS/LDS DiscoveryResponse (YAML) file and decodes its resources.
func loadResourceFile(file string) ([]cachetypes.Resource, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	resp, err := ugrpc.UnmarshalYAML[*discoveryservice.DiscoveryResponse](data)
	if err != nil {
		return nil, err
	}

	return decodeResources(resp.GetResources())
}
