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

package svc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func Test_ProxyHealthHandler_UpstreamServing(t *testing.T) {
	handler := newPongHandler()

	// Start upstream service with health
	upstream, err := NewHealthService(Configuration{
		Address: "127.0.0.1:0",
	}, Readiness, handler)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	st := upstream.Start(ctx)

	conn, err := grpc.NewClient(st.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	// Create proxyHealthHandler and verify it reports Healthy when upstream is serving
	h := &proxyHealthHandler{conn: cache.Static(conn)}

	state := h.Health(context.Background())
	require.Equal(t, Healthy, state)

	cancel()
	require.NoError(t, st.Wait())
}

func Test_ProxyHealthHandler_UpstreamStopped(t *testing.T) {
	handler := newPongHandler()

	// Start upstream service with health
	upstream, err := NewHealthService(Configuration{
		Address: "127.0.0.1:0",
	}, Readiness, handler)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	st := upstream.Start(ctx)

	conn, err := grpc.NewClient(st.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	h := &proxyHealthHandler{conn: cache.Static(conn)}

	// Verify healthy first
	state := h.Health(context.Background())
	require.Equal(t, Healthy, state)

	// Stop upstream
	cancel()
	require.NoError(t, st.Wait())

	// Health check should now return Unhealthy
	state = h.Health(context.Background())
	require.Equal(t, Unhealthy, state)
}

func Test_ProxyHealthHandler_ConnectionError(t *testing.T) {
	// Connect to an address where nothing is listening
	conn, err := grpc.NewClient("127.0.0.1:1", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	h := &proxyHealthHandler{conn: cache.Static(conn)}

	state := h.Health(context.Background())
	require.Equal(t, Unhealthy, state)
}

func Test_ProxyHealthHandler_Name(t *testing.T) {
	h := &proxyHealthHandler{}
	require.Equal(t, "proxy-upstream", h.Name())
}
