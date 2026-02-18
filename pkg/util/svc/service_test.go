//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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
	"fmt"
	"net"
	goHttp "net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"

	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

func Test_Service(t *testing.T) {
	h, err := NewHealthService(Configuration{
		Address: "127.0.0.1:0",
	}, Readiness)
	require.NoError(t, err)

	other, err := NewService(Configuration{
		Address: "127.0.0.1:0",
	})
	require.NoError(t, err)

	ctx, c := context.WithCancel(context.Background())
	defer c()

	st := h.Start(ctx)

	othStart := other.StartWithHealth(ctx, h)

	healthConn, err := grpc.NewClient(st.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	defer healthConn.Close()

	otherConn, err := grpc.NewClient(othStart.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	defer otherConn.Close()

	cl := pbHealth.NewHealthClient(healthConn)

	_, err = cl.Check(context.Background(), &pbHealth.HealthCheckRequest{})
	require.NoError(t, err)

	ol := pbHealth.NewHealthClient(otherConn)

	_, err = ol.Check(context.Background(), &pbHealth.HealthCheckRequest{})
	require.Error(t, err)

	c()

	require.NoError(t, st.Wait())
}

func Test_Service_Connections(t *testing.T) {
	handler := newPongHandler()

	t.Run("HTTP UNIX", func(t *testing.T) {
		dir := t.TempDir()

		h, err := NewService(Configuration{
			Address: "127.0.0.1:0",
			Unix:    filepath.Join(dir, "data.sock"),
			Gateway: &ConfigurationGateway{
				Address: "127.0.0.1:0",
				Unix:    filepath.Join(dir, "data2.sock"),
			},
		}, handler)
		require.NoError(t, err)

		ctx, c := context.WithCancel(context.Background())
		defer c()

		st := h.Start(ctx)

		conn, err := h.Dial()
		require.NoError(t, err)

		cl := pbPongV1.NewPongV1Client(conn)

		_, err = cl.Ping(context.Background(), &pbSharedV1.Empty{})
		require.NoError(t, err)

		_, err = ugrpc.Get[*pbPongV1.PongV1PingResponse](context.Background(), operatorHTTP.NewHTTPClient(
			operatorHTTP.WithTransport(func(in *goHttp.Transport) {
				in.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return net.Dial("unix", st.HTTPUnix())
				}
			}),
		), "http://unix/_integration/pong/v1/ping").WithCode(goHttp.StatusOK).Get()
		require.NoError(t, err)

		c()

		require.NoError(t, st.Wait())
	})

	t.Run("UNIX", func(t *testing.T) {
		dir := t.TempDir()

		h, err := NewService(Configuration{
			Address: "127.0.0.1:0",
			Unix:    filepath.Join(dir, "data.sock"),
			Gateway: &ConfigurationGateway{
				Address: "127.0.0.1:0",
			},
		}, handler)
		require.NoError(t, err)

		ctx, c := context.WithCancel(context.Background())
		defer c()

		st := h.Start(ctx)

		conn, err := h.Dial()
		require.NoError(t, err)

		cl := pbPongV1.NewPongV1Client(conn)

		_, err = cl.Ping(context.Background(), &pbSharedV1.Empty{})
		require.NoError(t, err)

		_, err = ugrpc.Get[*pbPongV1.PongV1PingResponse](context.Background(), operatorHTTP.NewHTTPClient(), fmt.Sprintf("http://%s/_integration/pong/v1/ping", st.HTTPAddress())).WithCode(goHttp.StatusOK).Get()
		require.NoError(t, err)

		c()

		require.NoError(t, st.Wait())
	})

	t.Run("TCP", func(t *testing.T) {
		h, err := NewService(Configuration{
			Address: "127.0.0.1:0",
			Gateway: &ConfigurationGateway{
				Address: "127.0.0.1:0",
			},
		}, handler)
		require.NoError(t, err)

		ctx, c := context.WithCancel(context.Background())
		defer c()

		st := h.Start(ctx)

		conn, err := h.Dial()
		require.NoError(t, err)

		cl := pbPongV1.NewPongV1Client(conn)

		_, err = cl.Ping(context.Background(), &pbSharedV1.Empty{})
		require.NoError(t, err)

		_, err = ugrpc.Get[*pbPongV1.PongV1PingResponse](context.Background(), operatorHTTP.NewHTTPClient(), fmt.Sprintf("http://%s/_integration/pong/v1/ping", st.HTTPAddress())).WithCode(goHttp.StatusOK).Get()
		require.NoError(t, err)

		c()

		require.NoError(t, st.Wait())
	})

}
