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

package v1

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pbPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1/definition"
	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Server(t *testing.T, ctx context.Context, services ...Service) svc.ServiceStarter {
	h, err := New(services...)
	require.NoError(t, err)

	local, err := svc.NewService(svc.Configuration{
		Address: "127.0.0.1:0",
		Gateway: &svc.ConfigurationGateway{
			Address: "127.0.0.1:0",
		},
	}, h)
	require.NoError(t, err)

	return local.Start(ctx)
}

func Client(t *testing.T, ctx context.Context, services ...Service) pbPongV1.PongV1Client {
	start := Server(t, ctx, services...)

	client := tgrpc.NewGRPCClient(t, ctx, pbPongV1.NewPongV1Client, start.Address())

	return client
}

func Test_Ping(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	client := Client(t, ctx)

	r1, err := client.Ping(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	r2, err := client.Ping(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)

	require.True(t, r2.GetTime().AsTime().After(r1.GetTime().AsTime()))
}

func Test_Ping_HTTP(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	server := Server(t, ctx)

	client := operatorHTTP.NewHTTPClient()

	resp := ugrpc.Get[*pbPongV1.PongV1PingResponse](ctx, client, fmt.Sprintf("http://%s/_integration/pong/v1/ping", server.HTTPAddress()))

	_, err := resp.WithCode(http.StatusOK).Get()
	require.NoError(t, err)
}

func Test_Services(t *testing.T) {
	ctx, c := context.WithCancel(context.Background())
	defer c()

	v2 := Service{
		Name:    "A",
		Version: "V2",
		Enabled: true,
	}

	v3 := Service{
		Name:    "A",
		Version: "V3",
		Enabled: false,
	}

	client := Client(t, ctx,
		v3,
		v2,
	)

	r1, err := client.Services(ctx, &pbSharedV1.Empty{})
	require.NoError(t, err)

	require.Len(t, r1.GetServices(), 2)

	require.NotNil(t, r1.GetServices()[0])
	require.NotNil(t, r1.GetServices()[1])

	require.EqualValues(t, v2.asService(), r1.GetServices()[0])
	require.EqualValues(t, v3.asService(), r1.GetServices()[1])
}
